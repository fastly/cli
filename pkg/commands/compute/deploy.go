package compute

import (
	"crypto/sha512"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/undo"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/kennygrant/sanitize"
)

type invalidResource int

const (
	defaultTopLevelDomain                 = "edgecompute.app"
	manageServiceBaseURL                  = "https://manage.fastly.com/configure/services/"
	resourceNone          invalidResource = iota
	resourceBoth
	resourceDomain
	resourceBackend
)

// DeployCommand deploys an artifact previously produced by build.
type DeployCommand struct {
	cmd.Base

	// NOTE: these are public so that the "publish" composite command can set the
	// values appropriately before calling the Exec() function.
	AcceptDefaults bool
	Backend        Backend
	Comment        cmd.OptionalString
	Domain         string
	Manifest       manifest.Data
	Path           string
	ServiceVersion cmd.OptionalServiceVersion
}

// Backend represents the configuration parameters for a backend
type Backend struct {
	Name           string
	Address        string
	OverrideHost   string
	Port           uint
	SSLSNIHostname string
	SetupConfig    bool
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = globals
	c.Manifest.File.SetOutput(c.Globals.Output)
	c.Manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute@Edge service")

	// NOTE: when updating these flags, be sure to update the composite command:
	// `compute publish`.
	c.RegisterServiceIDFlag(&c.Manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Action:   c.ServiceVersion.Set,
		Dst:      &c.ServiceVersion.Value,
		Optional: true,
	})
	c.CmdClause.Flag("accept-defaults", "Accept default values for all prompts and perform deploy non-interactively").BoolVar(&c.AcceptDefaults)
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.Domain)
	c.CmdClause.Flag("path", "Path to package").Short('p').StringVar(&c.Path)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	// The first thing we want to do is validate that a package has been built.
	// There is no point prompting a user for info if we know we're going to
	// fail any way because the user didn't build a package first.
	name, source := c.Manifest.Name()
	path, err := pkgPath(c.Path, name, source)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path":   c.Path,
			"Name":   name,
			"Source": source,
		})
		return err
	}
	if err := validate(path); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path": path,
		})
		return err
	}

	var (
		domain          string
		backends        []Backend
		backendsMissing []manifest.Mapper
		invalidService  bool
		invalidType     invalidResource
		version         *fastly.Version
	)

	serviceID, sidSrc := c.Manifest.ServiceID()
	if sidSrc == manifest.SourceUndefined {
		text.Break(out)
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")
		text.Break(out)

		backends, err = configureBackends(c, backends, c.Manifest.File.Setup.Backends, out, in)
		if err != nil {
			return err
		}

		domain, err = cfgDomain(c, defaultTopLevelDomain, out, in, validateDomain)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Domain":           c.Domain,
				"Domain (default)": defaultTopLevelDomain,
			})
			return err
		}

		text.Break(out)
	} else {
		version, err = c.ServiceVersion.Parse(serviceID, c.Globals.Client)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID": serviceID,
			})
			return err
		}

		// Unlike other CLI commands that are a direct mapping to an API endpoint,
		// the compute deploy command is a composite of behaviours, and so as we
		// already automatically activate a version we should autoclone without
		// requiring the user to explicitly provide an --autoclone flag.
		if version.Active || version.Locked {
			v, err := c.Globals.Client.CloneVersion(&fastly.CloneVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: version.Number,
			})
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Service ID":      serviceID,
					"Service Version": version.Number,
				})
				return fmt.Errorf("error cloning service version: %w", err)
			}
			if c.Globals.Flag.Verbose {
				msg := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned. Now operating on version %d.", version.Number, v.Number)
				text.Break(out)
				text.Output(out, msg)
				text.Break(out)
			}
			version = v
		}

		// We define the `ok` variable so that the following call to
		// `validateservice` will be able to shadow `invalidType`.
		var ok bool

		// Because a service_id exists in the fastly.toml doesn't mean it's valid
		// e.g. it could be missing either a domain or backend resource. So we
		// check and allow the user to configure these settings before continuing.
		ok, invalidType, backendsMissing, err = validateService(serviceID, c.Globals.Client, version, c.Manifest.File.Setup.Backends)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID":      serviceID,
				"Service Version": version.Number,
			})
			return err
		}

		if !ok {
			invalidService = true

			text.Output(out, "Service '%s' is missing required domain or backend(s). These must be added before the Compute@Edge service can be deployed.", serviceID)
			text.Break(out)

			switch invalidType {
			case resourceBoth:
				domain, err = cfgDomain(c, defaultTopLevelDomain, out, in, validateDomain)
				if err != nil {
					c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
						"Domain":           c.Domain,
						"Domain (default)": defaultTopLevelDomain,
					})
					return err
				}
				backends, err = configureBackends(c, backends, backendsMissing, out, in)
				if err != nil {
					return err
				}
			case resourceDomain:
				domain, err = cfgDomain(c, defaultTopLevelDomain, out, in, validateDomain)
				if err != nil {
					c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
						"Domain":           c.Domain,
						"Domain (default)": defaultTopLevelDomain,
					})
					return err
				}
			case resourceBackend:
				backends, err = configureBackends(c, backends, backendsMissing, out, in)
				if err != nil {
					return err
				}
			}

			text.Break(out)
		}
	}

	var (
		progress text.Progress
		desc     string
	)

	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	undoStack := undo.NewStack()

	defer func(errLog errors.LogInterface) {
		if err != nil {
			errLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
		undoStack.RunIfError(out, err)
	}(c.Globals.ErrLog)

	if sidSrc == manifest.SourceUndefined {
		// There is no service and so we'll do a one time creation of the service
		// and the associated domain/backend(s) and store the Service ID within the
		// manifest. On subsequent runs of the deploy subcommand we'll skip the
		// service/domain/backend creation.
		//
		// NOTE: we're shadowing the `version` and `serviceID` variable.
		version, serviceID, err = createService(progress, c.Globals.Client, name, desc)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Name":        name,
				"Description": desc,
			})
			return err
		}

		undoStack.Push(func() error {
			clearServiceID := ""
			return updateManifestServiceID(&c.Manifest.File, manifest.Filename, nil, clearServiceID)
		})

		undoStack.Push(func() error {
			return c.Globals.Client.DeleteService(&fastly.DeleteServiceInput{
				ID: serviceID,
			})
		})

		// We can't create the domain/backend earlier in the logic flow as it
		// requires the use of a text.Progress which overwrites the current line
		// (i.e. it would cause any text prompts to be hidden) and so we prompt for
		// as much information as possible at the top of the Exec function. After
		// we have all the information, then we proceed with the creation of resources.
		err = createDomain(progress, c.Globals.Client, serviceID, version.Number, domain, undoStack)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Domain":          domain,
				"Service ID":      serviceID,
				"Service Version": version.Number,
			})
			return err
		}

		for _, backend := range backends {
			err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, undoStack)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Accept defaults": c.AcceptDefaults,
					"Service ID":      serviceID,
					"Service Version": version.Number,
				})
				return err
			}
		}

		err = updateManifestServiceID(&c.Manifest.File, manifest.Filename, progress, serviceID)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID": serviceID,
			})
			return err
		}
	}

	// If the user has specified a Service ID, then we validate it has the
	// required resources, and if it's invalid we'll drop into the following code
	// block to ensure we only create the resources that are missing.
	if invalidService {
		switch invalidType {
		case resourceBoth:
			err = createDomain(progress, c.Globals.Client, serviceID, version.Number, domain, undoStack)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Domain":          domain,
					"Service ID":      serviceID,
					"Service Version": version.Number,
				})
				return err
			}
			for _, backend := range backends {
				err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, undoStack)
				if err != nil {
					c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
						"Accept defaults": c.AcceptDefaults,
						"Service ID":      serviceID,
						"Service Version": version.Number,
					})
					return err
				}
			}
		case resourceDomain:
			err = createDomain(progress, c.Globals.Client, serviceID, version.Number, domain, undoStack)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Domain":          domain,
					"Service ID":      serviceID,
					"Service Version": version.Number,
				})
				return err
			}
		case resourceBackend:
			for _, backend := range backends {
				err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, undoStack)
				if err != nil {
					c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
						"Accept defaults": c.AcceptDefaults,
						"Service ID":      serviceID,
						"Service Version": version.Number,
					})
					return err
				}
			}
		}
	}

	cont, err := pkgCompare(c.Globals.Client, serviceID, version.Number, path, progress, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path":            path,
			"Service ID":      serviceID,
			"Service Version": version.Number,
		})
		return err
	}
	if !cont {
		return nil
	}

	err = pkgUpload(progress, c.Globals.Client, serviceID, version.Number, path)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path":            path,
			"Service ID":      serviceID,
			"Service Version": version.Number,
		})
		return err
	}

	if c.Comment.WasSet {
		_, err = c.Globals.Client.UpdateVersion(&fastly.UpdateVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: version.Number,
			Comment:        &c.Comment.Value,
		})

		if err != nil {
			return fmt.Errorf("error setting comment for service version %d: %w", version.Number, err)
		}
	}

	progress.Step("Activating version...")

	_, err = c.Globals.Client.ActivateVersion(&fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": version.Number,
		})
		return fmt.Errorf("error activating version: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))

	domains, err := c.Globals.Client.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err == nil {
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", domains[0].Name))
	}

	text.Success(out, "Deployed package (service %s, version %v)", serviceID, version.Number)
	return nil
}

// configureBackends determines whether multiple backends need to be configured
// or a singular backend based on existence of the fastly.toml [setup] table.
func configureBackends(c *DeployCommand, bs []Backend, predefinedBackends []manifest.Mapper, out io.Writer, in io.Reader) ([]Backend, error) {
	if len(predefinedBackends) > 0 {
		for _, backend := range predefinedBackends {
			b, err := cfgSetupBackend(backend, c, out, in, validateBackend)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Backends (predefined)": predefinedBackends,
				})
				return nil, err
			}
			bs = append(bs, b)
		}
	} else {
		backends, err := cfgBackends(c, out, in, validateBackend)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Accept defaults": c.AcceptDefaults,
			})
			return nil, err
		}
		bs = append(bs, backends...)
	}
	return bs, nil
}

// pkgPath generates a path that points to a package tar inside the pkg
// directory if the `path` flag was not set by the user.
func pkgPath(path string, name string, source manifest.Source) (string, error) {
	if path == "" {
		if source == manifest.SourceUndefined {
			return "", errors.RemediationError{
				Inner:       fmt.Errorf("error reading package manifest"),
				Remediation: "Run `fastly compute init` to ensure a correctly configured manifest. See more at https://developer.fastly.com/reference/fastly-toml/",
			}
		}

		path = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(name)))
		return path, nil
	}

	return path, nil
}

// validateService checks if the service version has a domain and backend
// defined, and in the case of the [setup] configuration it will validate the
// backends defined in the config exist in the given service.
//
// The use of `resourceNone` as a `invalidResource` enum type is to represent a
// situation where an error occurred and so we were unable to identify whether
// a domain or backend (or both) were missing from the given service.
func validateService(serviceID string, client api.Interface, version *fastly.Version, predefinedBackends []manifest.Mapper) (bool, invalidResource, []manifest.Mapper, error) {
	var missingBackends []manifest.Mapper

	// We require at least one backend but if the project fastly.toml contains
	// a [setup] with backends defined, then we'll use that as our requirement.
	requiredBackends := 1
	definedBackends := len(predefinedBackends)
	if definedBackends > 0 {
		requiredBackends = definedBackends
	}

	err := checkServiceID(serviceID, client, version)
	if err != nil {
		return false, resourceNone, missingBackends, err
	}

	domains, err := client.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		return false, resourceNone, missingBackends, fmt.Errorf("error fetching service domains: %w", err)
	}

	backends, err := client.ListBackends(&fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		return false, resourceNone, missingBackends, fmt.Errorf("error fetching service backends: %w", err)
	}

	// At a minimum we need the number of backends in the existing service to
	// match requiredBackends (and at least one existing domain).
	{
		ld := len(domains)
		lb := len(backends)
		if ld == 0 && lb < requiredBackends {
			return false, resourceBoth, missingBackends, nil
		}
		if ld == 0 {
			return false, resourceDomain, missingBackends, nil
		}
		if lb < requiredBackends {
			return false, resourceBackend, missingBackends, nil
		}
	}

	// Error descriptions used if we need to handle [setup] configuration
	innerErr := fmt.Errorf("error parsing the [[setup.backends]] configuration")
	remediation := "Check the fastly.toml configuration for a missing or invalid '%s' field."

	for _, backend := range predefinedBackends {
		if _, exists := backend["address"]; exists {
			addr, ok := backend["address"].(string)
			if !ok || addr == "" {
				return false, resourceNone, missingBackends, backendRemediationError("address", remediation, innerErr)
			}
			match := false
			for _, bs := range backends {
				if bs.Address == addr {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
		}

		if _, exists := backend["name"]; exists {
			name, ok := backend["name"].(string)
			if !ok {
				return false, resourceNone, missingBackends, backendRemediationError("name", remediation, innerErr)
			}
			match := false
			for _, bs := range backends {
				if bs.Name == name {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
		}

		if _, exists := backend["port"]; exists {
			port, ok := backend["port"].(int64)
			if !ok {
				return false, resourceNone, missingBackends, backendRemediationError("port", remediation, innerErr)
			}
			match := false
			for _, bs := range backends {
				if bs.Port == uint(port) {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
		}
	}

	return true, resourceNone, missingBackends, nil
}

// checkServiceID validates the given Service ID maps to a real service.
func checkServiceID(serviceID string, client api.Interface, version *fastly.Version) error {
	_, err := client.GetService(&fastly.GetServiceInput{
		ID: serviceID,
	})
	if err != nil {
		return fmt.Errorf("error fetching service details: %w", err)
	}
	return nil
}

// createService creates a service to associate with the compute package.
func createService(progress text.Progress, client api.Interface, name string, desc string) (*fastly.Version, string, error) {
	progress.Step("Creating service...")

	service, err := client.CreateService(&fastly.CreateServiceInput{
		Name:    name,
		Type:    "wasm",
		Comment: desc,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Valid values for 'type' are: 'vcl'") {
			return nil, "", errors.RemediationError{
				Inner:       fmt.Errorf("error creating service: you do not have the Compute@Edge feature flag enabled on your Fastly account"),
				Remediation: "For more help with this error see fastly.help/cli/ecp-feature",
			}
		}
		return nil, "", fmt.Errorf("error creating service: %w", err)
	}

	version := &fastly.Version{Number: 1}
	serviceID := service.ID

	return version, serviceID, nil
}

// updateManifestServiceID updates the Service ID in the manifest.
//
// There are two scenarios where this function is called. The first is when we
// have a Service ID to insert into the manifest. The other is when there is an
// error in the deploy flow, and for which the Service ID will be set to an
// empty string (otherwise the service itself will be deleted while the
// manifest will continue to hold a reference to it).
func updateManifestServiceID(m *manifest.File, manifestFilename string, progress text.Progress, serviceID string) error {
	if err := m.Read(manifestFilename); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	if progress != nil {
		fmt.Fprintf(progress, "Setting service ID in manifest to %q...\n", serviceID)
	}

	m.ServiceID = serviceID

	if err := m.Write(manifestFilename); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	return nil
}

// pkgCompare compares the local package hashsum against the existing service
// package version and exits early with message if identical.
func pkgCompare(client api.Interface, serviceID string, version int, path string, progress text.Progress, out io.Writer) (bool, error) {
	p, err := client.GetPackage(&fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})

	if err == nil {
		hashSum, err := getHashSum(path)
		if err != nil {
			return false, fmt.Errorf("error getting package hashsum: %w", err)
		}

		if hashSum == p.Metadata.HashSum {
			progress.Done()
			text.Info(out, "Skipping package deployment, local and service version are identical. (service %v, version %v) ", serviceID, version)
			return false, nil
		}
	}

	return true, nil
}

// pkgUpload uploads the package to the specified service and version.
func pkgUpload(progress text.Progress, client api.Interface, serviceID string, version int, path string) error {
	progress.Step("Uploading package...")

	_, err := client.UpdatePackage(&fastly.UpdatePackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		PackagePath:    path,
	})

	if err != nil {
		return fmt.Errorf("error uploading package: %w", err)
	}

	return nil
}

// cfgDomain configures the domain value.
func cfgDomain(c *DeployCommand, def string, out io.Writer, in io.Reader, f validator) (string, error) {
	if c.Domain != "" {
		return c.Domain, nil
	}

	rand.Seed(time.Now().UnixNano())
	defaultDomain := fmt.Sprintf("%s.%s", petname.Generate(3, "-"), def)

	var (
		domain string
		err    error
	)
	if !c.AcceptDefaults {
		domain, err = text.Input(out, fmt.Sprintf("Domain: [%s] ", defaultDomain), in, f)
		if err != nil {
			return "", fmt.Errorf("error reading input %w", err)
		}
	}

	if domain == "" {
		return defaultDomain, nil
	}
	return domain, nil
}

// cfgSetupBackend configures the backend address and its port number values
// with values provided by the fastly.toml [setup] configuration.
func cfgSetupBackend(backend manifest.Mapper, c *DeployCommand, out io.Writer, in io.Reader, v validator) (Backend, error) {
	var (
		addr        string
		b           Backend
		defaultAddr string
		err         error
		name        string
		ok          bool
		port        uint
		prompt      string
	)

	b.SetupConfig = true

	innerErr := fmt.Errorf("error parsing the [[setup.backends]] configuration")
	remediation := "Check the fastly.toml configuration for a missing or invalid '%s' field."

	// ADDRESS DEFAULT (REQUIRED)
	{
		if _, ok = backend["address"]; ok {
			addr, ok = backend["address"].(string)
			if !ok {
				return b, backendRemediationError("address", remediation, innerErr)
			}
		} else {
			return b, backendRemediationError("address", remediation, innerErr)
		}
	}

	// NAME DEFAULT
	{
		if _, ok = backend["name"]; ok {
			name, ok = backend["name"].(string)
			if !ok {
				return b, backendRemediationError("name", remediation, innerErr)
			}
		}
		b.Name = name
		if name == "" {
			b.Name = genBackendName(addr)
		}
	}

	// PROMPT DEFAULT
	{
		if _, ok := backend["prompt"]; ok {
			prompt, ok = backend["prompt"].(string)
			if !ok {
				return b, backendRemediationError("prompt", remediation, innerErr)
			}
		}
		// If no prompt text is provided by the [setup] configuration, then we'll
		// default to using the name of the backend as the prompt text.
		if prompt == "" {
			prompt = fmt.Sprintf("Backend for '%s'", b.Name)
		}
	}

	// PORT DEFAULT
	{
		if _, ok = backend["port"]; ok {
			p, ok := backend["port"].(int64)
			if !ok {
				return b, backendRemediationError("port", remediation, innerErr)
			}
			port = uint(p)
		}
		if port == 0 {
			port = 80
		}
	}

	// PROMPT USER INTERACTIVELY FOR ADDRESS AND PORT...

	if !c.AcceptDefaults {
		defaultAddr = fmt.Sprintf(": [%s] ", addr)
		b.Address, err = text.Input(out, fmt.Sprintf("%s%s ", prompt, defaultAddr), in, v)
		if err != nil {
			return b, fmt.Errorf("error reading input %w", err)
		}
	}
	if b.Address == "" {
		b.Address = addr
	}

	if !c.AcceptDefaults {
		input, err := text.Input(out, fmt.Sprintf("Backend port number: [%d] ", port), in)
		if err != nil {
			return b, fmt.Errorf("error reading input %w", err)
		}
		if input != "" {
			i, err := strconv.Atoi(input)
			if err != nil {
				text.Warning(out, fmt.Sprintf("error converting input, using default port number (%d)", port))
			} else {
				port = uint(i)
			}
		}
	}
	b.Port = port

	setBackendHost(&b)

	return b, nil
}

// backendRemediationError reduces the boilerplate of serving a remediation
// error whose only difference is the field it applies to.
func backendRemediationError(field string, remediation string, err error) error {
	return errors.RemediationError{
		Inner:       err,
		Remediation: fmt.Sprintf(remediation, field),
	}
}

// cfgBackends configures multiple backends based on prompted input from the user.
//
// NOTE: If `--accept-defaults` is set, then create a single "originless" backend.
func cfgBackends(c *DeployCommand, out io.Writer, in io.Reader, f validator) (backends []Backend, err error) {
	if c.AcceptDefaults {
		backend := createOriginlessBackend()
		backends = append(backends, backend)
		return backends, nil
	}

	for {
		var backend Backend

		if backend.Name == "" {
			backend.Name = backend.Address
		}

		backend.Address, err = text.Input(out, "Backend (hostname or IP address, or leave blank to stop adding backends): ", in, f)
		if err != nil {
			return backends, fmt.Errorf("error reading input %w", err)
		}

		// This block short-circuits the endless loop
		if backend.Address == "" {
			if len(backends) == 0 {
				backend := createOriginlessBackend()
				backends = append(backends, backend)
				return backends, nil
			}
			return backends, nil
		}

		input, err := text.Input(out, "Backend port number: [80] ", in)
		if err != nil {
			return backends, fmt.Errorf("error reading input %w", err)
		}

		portnumber := 80
		if input != "" {
			portnumber, err = strconv.Atoi(input)
			if err != nil {
				text.Warning(out, "error converting input, using default port number (80)")
				portnumber = 80
			}
		}

		backend.Port = uint(portnumber)

		if backend.Name == "" {
			backend.Name, err = text.Input(out, "Backend name: ", in)
			if err != nil {
				return backends, fmt.Errorf("error reading input %w", err)
			}
			if backend.Name == "" {
				backend.Name = genBackendName(backend.Address)
			}
		}

		setBackendHost(&backend)

		backends = append(backends, backend)
	}
}

// createOriginlessBackend returns a Backend instance configured to the
// localhost settings expected of an 'originless' backend.
func createOriginlessBackend() (b Backend) {
	b.Name = "originless"
	b.Address = "127.0.0.1"
	b.Port = uint(80)
	return b
}

// genBackendName normalises a given name by replacing any period or hyphen
// characters with an underscore.
func genBackendName(name string) string {
	normalise := "_"
	r := strings.NewReplacer(".", normalise, "-", normalise)
	return r.Replace(name)
}

// setBackendHost sets two fields: OverrideHost and SSLSNIHostname.
func setBackendHost(b *Backend) {
	// By default set the override_host and ssl_sni_hostname properties of the
	// Backend VCL object to the hostname unless the given input is an IP.
	b.OverrideHost = b.Address
	if _, err := net.LookupAddr(b.Address); err == nil {
		b.OverrideHost = ""
	}
	if b.OverrideHost != "" {
		b.SSLSNIHostname = b.OverrideHost
	}
}

// createDomain creates the given domain and handle unrolling the stack in case
// of an error (i.e. will ensure the domain is deleted if there is an error).
func createDomain(progress text.Progress, client api.Interface, serviceID string, version int, domain string, undoStack undo.Stacker) error {
	progress.Step("Creating domain...")

	undoStack.Push(func() error {
		return client.DeleteDomain(&fastly.DeleteDomainInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           domain,
		})
	})

	_, err := client.CreateDomain(&fastly.CreateDomainInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           domain,
	})
	if err != nil {
		return fmt.Errorf("error creating domain: %w", err)
	}

	return nil
}

// createBackend creates the given domain and handle unrolling the stack in case
// of an error (i.e. will ensure the backend is deleted if there is an error).
func createBackend(progress text.Progress, client api.Interface, serviceID string, version int, backend Backend, undoStack undo.Stacker) error {
	// We don't display the fact we're creating a backend when it's for an
	// originless purpose as the user shouldn't have to know about this detail.
	originless := backend.Name == "originless" && backend.Address == "127.0.0.1"
	if !originless {
		progress.Step(fmt.Sprintf("Creating backend '%s' (host: %s, port: %d)...", backend.Name, backend.Address, backend.Port))
	}

	undoStack.Push(func() error {
		return client.DeleteBackend(&fastly.DeleteBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           backend.Address,
		})
	})

	_, err := client.CreateBackend(&fastly.CreateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           backend.Name,
		Address:        backend.Address,
		Port:           backend.Port,
		OverrideHost:   backend.OverrideHost,
		SSLSNIHostname: backend.SSLSNIHostname,
	})
	if err != nil {
		if originless {
			return fmt.Errorf("error configuring the service: %w", err)
		} else {
			return fmt.Errorf("error creating backend: %w", err)
		}
	}

	return nil
}

func getHashSum(path string) (hash string, err error) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the filepath variable.
	/* #nosec */
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type validator func(input string) error

func validateBackend(input string) error {
	var isHost bool
	if _, err := net.LookupHost(input); err == nil {
		isHost = true
	}
	var isAddr bool
	if _, err := net.LookupAddr(input); err == nil {
		isAddr = true
	}
	isEmpty := input == ""
	if !isEmpty && !isHost && !isAddr {
		return fmt.Errorf(`must be a valid hostname, IPv4, or IPv6 address`)
	}
	return nil
}

func validateDomain(input string) error {
	if input == "" {
		return nil
	}
	if !domainNameRegEx.MatchString(input) {
		return fmt.Errorf("must be valid domain name")
	}
	return nil
}
