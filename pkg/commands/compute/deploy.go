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
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/kennygrant/sanitize"
)

const (
	defaultTopLevelDomain = "edgecompute.app"
	manageServiceBaseURL  = "https://manage.fastly.com/configure/services/"
)

// PackageSizeLimit describes the package size limit in bytes (currently 50mb)
// https://docs.fastly.com/products/compute-at-edge-billing-and-resource-limits#resource-limits
var PackageSizeLimit int64 = 50000000

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
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data, data manifest.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = globals
	c.Manifest = data
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
	c.CmdClause.Flag("name", "Package name").StringVar(&c.Manifest.Flag.Name)
	c.CmdClause.Flag("path", "Path to package").Short('p').StringVar(&c.Path)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	serviceID, sidSrc := c.Manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, sidSrc, out)
	}

	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	// Alias' for otherwise long definitions
	errLog := c.Globals.ErrLog
	verbose := c.Globals.Verbose()
	apiClient := c.Globals.Client

	// VALIDATE PACKAGE...

	pkgName, pkgPath, err := validatePackage(c.Manifest, c.Path, errLog)
	if err != nil {
		return err
	}

	// SERVICE MANAGEMENT...

	var (
		newService     bool
		serviceVersion *fastly.Version
	)

	if sidSrc == manifest.SourceUndefined {
		newService = true
		serviceID, serviceVersion, err = manageNoServiceIDFlow(c.AcceptDefaults, in, out, verbose, apiClient, pkgName, errLog, &c.Manifest.File)
		if err != nil {
			return err
		}
		if serviceID == "" {
			// The user said NO to creating a service when prompted.
			return nil
		}
	} else {
		serviceVersion, err = manageExistingServiceFlow(serviceID, c.ServiceVersion, apiClient, verbose, out, errLog)
		if err != nil {
			return err
		}
	}

	// RESOURCE VALIDATION...

	// We only check the Service ID is valid when handling an existing service.
	if !newService {
		err = checkServiceID(serviceID, apiClient, serviceVersion)
		if err != nil {
			errLogService(errLog, err, serviceID, serviceVersion.Number)
			return err
		}
	}

	// Because a service_id exists in the fastly.toml doesn't mean it's valid
	// e.g. it could be missing required resources such as a domain or backend.
	// We check and allow the user to configure these settings before continuing.

	hasDomain, err := serviceHasDomain(serviceID, serviceVersion.Number, apiClient)
	if err != nil {
		errLogService(errLog, err, serviceID, serviceVersion.Number)
		return err
	}

	availableBackends, err := apiClient.ListBackends(&fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	})
	if err != nil {
		errLogService(errLog, err, serviceID, serviceVersion.Number)
		return fmt.Errorf("error fetching service backends: %w", err)
	}

	var (
		backendsToCreate    []Backend
		domainToCreate      string
		hasRequiredBackends bool
		missingBackends     []manifest.Mapper
	)

	predefinedBackends := c.Manifest.File.Setup.Backends
	if len(predefinedBackends) > 0 {
		missingBackends, err = serviceMissingConfiguredBackends(serviceID, serviceVersion.Number, availableBackends, predefinedBackends)
		if err != nil {
			errLogService(errLog, err, serviceID, serviceVersion.Number)
			return err
		}
		if len(missingBackends) == 0 {
			hasRequiredBackends = true
		}
	} else if len(availableBackends) > 0 {
		hasRequiredBackends = true
	}

	// RESOURCE CONFIGURATION...

	if !hasDomain || !hasRequiredBackends {
		if !c.AcceptDefaults {
			text.Output(out, "Service '%s' is missing required resources. These must be added before the Compute@Edge service can be deployed. Please ensure your fastly.toml configuration reflects any manual changes made via manage.fastly.com, otherwise follow the prompts to create the required resources.", serviceID)
			text.Break(out)
		}
	}

	if !hasDomain {
		domainToCreate, err = configureDomain(c, defaultTopLevelDomain, out, in, validateDomain)
		if err != nil {
			errLog.AddWithContext(err, map[string]interface{}{
				"Domain":           c.Domain,
				"Domain (default)": defaultTopLevelDomain,
			})
			return err
		}
	}

	if !hasRequiredBackends {
		backendsToCreate, err = configureBackends(c, backendsToCreate, missingBackends, out, in)
		if err != nil {
			return err
		}
	}

	text.Break(out)

	// RESOURCE CREATION...

	progress := text.NewProgress(out, c.Globals.Verbose())
	undoStack := undo.NewStack()

	defer func(errLog errors.LogInterface, progress text.Progress) {
		if err != nil {
			errLog.Add(err)
			progress.Fail()
		}
		undoStack.RunIfError(out, err)
	}(errLog, progress)

	if !hasDomain && domainToCreate != "" {
		err = createDomain(progress, apiClient, serviceID, serviceVersion.Number, domainToCreate, undoStack)
		if err != nil {
			errLog.AddWithContext(err, map[string]interface{}{
				"Domain":          domainToCreate,
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
	}

	for _, backend := range backendsToCreate {
		err = createBackend(progress, apiClient, serviceID, serviceVersion.Number, backend, undoStack)
		if err != nil {
			errLog.AddWithContext(err, map[string]interface{}{
				"Accept defaults": c.AcceptDefaults,
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
	}

	// PACKAGE PROCESSING...

	cont, err := pkgCompare(apiClient, serviceID, serviceVersion.Number, pkgPath, progress, out)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Package path":    pkgPath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}
	if !cont {
		return nil
	}

	err = pkgUpload(progress, apiClient, serviceID, serviceVersion.Number, pkgPath)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Package path":    pkgPath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	// SERVICE PROCESSING...

	if c.Comment.WasSet {
		_, err = apiClient.UpdateVersion(&fastly.UpdateVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion.Number,
			Comment:        &c.Comment.Value,
		})

		if err != nil {
			return fmt.Errorf("error setting comment for service version %d: %w", serviceVersion.Number, err)
		}
	}

	progress.Step("Activating version...")

	_, err = apiClient.ActivateVersion(&fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	})
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return fmt.Errorf("error activating version: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))

	domains, err := apiClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	})
	if err == nil {
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", domains[0].Name))
	}

	text.Success(out, "Deployed package (service %s, version %v)", serviceID, serviceVersion.Number)
	return nil
}

// validatePackage short-circuits the deploy command if the user hasn't first
// built a package to be deployed.
//
// NOTE: It also validates if the package size exceeds limit:
// https://docs.fastly.com/products/compute-at-edge-billing-and-resource-limits#resource-limits
func validatePackage(data manifest.Data, pathFlag string, errLog errors.LogInterface) (pkgName, pkgPath string, err error) {
	pkgName, source := data.Name()
	pkgPath, err = packagePath(pathFlag, pkgName, source)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Package path": pathFlag,
			"Package name": pkgName,
			"Source":       source,
		})
		return pkgName, pkgPath, err
	}
	pkgSize, err := packageSize(pkgPath)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Package path": pkgPath,
		})
		return pkgName, pkgPath, err
	}
	if pkgSize > PackageSizeLimit {
		return pkgName, pkgPath, errors.RemediationError{
			Inner:       fmt.Errorf("package size is too large (%d bytes)", pkgSize),
			Remediation: errors.PackageSizeRemediation,
		}
	}
	if err := validate(pkgPath); err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Package path": pkgPath,
			"Package size": pkgSize,
		})
		return pkgName, pkgPath, err
	}
	return pkgName, pkgPath, nil
}

// packagePath generates a path that points to a package tar inside the pkg
// directory if the `path` flag was not set by the user.
func packagePath(path string, name string, source manifest.Source) (string, error) {
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

// packageSize returns the size of the .tar.gz package.
func packageSize(path string) (size int64, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return size, err
	}
	return fi.Size(), nil
}

// manageNoServiceIDFlow handles creating a new service when no Service ID is found.
func manageNoServiceIDFlow(
	acceptDefaults bool,
	in io.Reader,
	out io.Writer,
	verbose bool,
	apiClient api.Interface,
	pkgName string,
	errLog errors.LogInterface,
	manifestFile *manifest.File) (serviceID string, serviceVersion *fastly.Version, err error) {

	if !acceptDefaults {
		text.Break(out)
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")
		text.Break(out)

		service, err := text.Input(out, "Create new service: [y/N] ", in)
		if err != nil {
			return serviceID, serviceVersion, fmt.Errorf("error reading input %w", err)
		}
		if service != "y" && service != "Y" {
			return serviceID, serviceVersion, nil
		}

		text.Break(out)
	}

	progress := text.NewProgress(out, verbose)

	// There is no service and so we'll do a one time creation of the service
	//
	// NOTE: we're shadowing the `serviceVersion` and `serviceID` variables.
	serviceID, serviceVersion, err = createService(apiClient, pkgName, progress)
	if err != nil {
		progress.Fail()
		errLog.AddWithContext(err, map[string]interface{}{
			"Package name": pkgName,
		})
		return serviceID, serviceVersion, err
	}

	progress.Done()

	err = updateManifestServiceID(manifestFile, manifest.Filename, serviceID)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return serviceID, serviceVersion, err
	}

	text.Break(out)
	return serviceID, serviceVersion, nil
}

// createService creates a service to associate with the compute package.
func createService(client api.Interface, name string, progress text.Progress) (string, *fastly.Version, error) {
	progress.Step("Creating service...")

	service, err := client.CreateService(&fastly.CreateServiceInput{
		Name: name,
		Type: "wasm",
	})
	if err != nil {
		if strings.Contains(err.Error(), "Valid values for 'type' are: 'vcl'") {
			return "", nil, errors.RemediationError{
				Inner:       fmt.Errorf("error creating service: you do not have the Compute@Edge feature flag enabled on your Fastly account"),
				Remediation: "For more help with this error see fastly.help/cli/ecp-feature",
			}
		}
		return "", nil, fmt.Errorf("error creating service: %w", err)
	}

	serviceID := service.ID
	serviceVersion := &fastly.Version{Number: 1}

	return serviceID, serviceVersion, nil
}

// updateManifestServiceID updates the Service ID in the manifest.
//
// There are two scenarios where this function is called. The first is when we
// have a Service ID to insert into the manifest. The other is when there is an
// error in the deploy flow, and for which the Service ID will be set to an
// empty string (otherwise the service itself will be deleted while the
// manifest will continue to hold a reference to it).
func updateManifestServiceID(m *manifest.File, manifestFilename string, serviceID string) error {
	if err := m.Read(manifestFilename); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	m.ServiceID = serviceID

	if err := m.Write(manifestFilename); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	return nil
}

// manageExistingServiceFlow clones service version if required.
func manageExistingServiceFlow(
	serviceID string,
	serviceVersionFlag cmd.OptionalServiceVersion,
	apiClient api.Interface,
	verbose bool,
	out io.Writer,
	errLog errors.LogInterface) (serviceVersion *fastly.Version, err error) {

	serviceVersion, err = serviceVersionFlag.Parse(serviceID, apiClient)
	if err != nil {
		errLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return serviceVersion, err
	}

	// Unlike other CLI commands that are a direct mapping to an API endpoint,
	// the compute deploy command is a composite of behaviours, and so as we
	// already automatically activate a version we should autoclone without
	// requiring the user to explicitly provide an --autoclone flag.
	if serviceVersion.Active || serviceVersion.Locked {
		clonedVersion, err := apiClient.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion.Number,
		})
		if err != nil {
			errLogService(errLog, err, serviceID, serviceVersion.Number)
			return serviceVersion, fmt.Errorf("error cloning service version: %w", err)
		}
		if verbose {
			msg := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned. Now operating on version %d.", serviceVersion.Number, clonedVersion.Number)
			text.Break(out)
			text.Output(out, msg)
			text.Break(out)
		}
		serviceVersion = clonedVersion
	}

	return serviceVersion, nil
}

// errLogService records the error, service id and version into the error log.
func errLogService(l errors.LogInterface, err error, sid string, sv int) {
	l.AddWithContext(err, map[string]interface{}{
		"Service ID":      sid,
		"Service Version": sv,
	})
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

// serviceMissingConfiguredBackends analyses the predefined [setup]
// configuration and returns a list of missing backends.
func serviceMissingConfiguredBackends(sid string, sv int, availableBackends []*fastly.Backend, predefinedBackends []manifest.Mapper) ([]manifest.Mapper, error) {
	var missingBackends []manifest.Mapper

	// Error descriptions used if we need to handle [setup] configuration
	innerErr := fmt.Errorf("error parsing the [[setup.backends]] configuration")
	remediation := "Check the fastly.toml configuration for a missing or invalid '%s' field."

	for _, backend := range predefinedBackends {
		var (
			hasAddress bool
			hasName    bool
			hasPort    bool

			addr string
			name string
			ok   bool
			port int64
		)

		// ACQUIRE CONFIG FIELDS...

		if _, exists := backend["address"]; exists {
			addr, ok = backend["address"].(string)
			if !ok || addr == "" {
				return missingBackends, backendRemediationError("address", remediation, innerErr)
			}
			hasAddress = true
		} else {
			return missingBackends, backendRemediationError("address", remediation, innerErr)
		}

		if _, exists := backend["name"]; exists {
			name, ok = backend["name"].(string)
			if !ok {
				return missingBackends, backendRemediationError("name", remediation, innerErr)
			}
			hasName = true
		}

		if _, exists := backend["port"]; exists {
			port, ok = backend["port"].(int64)
			if !ok {
				return missingBackends, backendRemediationError("port", remediation, innerErr)
			}
			hasPort = true
		}

		// VALIDATE CONFIG FIELDS...

		if hasAddress && hasName && hasPort {
			match := false
			for _, bs := range availableBackends {
				if bs.Address == addr && bs.Name == name && bs.Port == uint(port) {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
			continue
		}

		if hasAddress && hasName {
			match := false
			for _, bs := range availableBackends {
				if bs.Address == addr && bs.Name == name {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
			continue
		}

		if hasAddress && hasPort {
			match := false
			for _, bs := range availableBackends {
				if bs.Address == addr && bs.Port == uint(port) {
					match = true
					break
				}
			}
			if !match {
				missingBackends = append(missingBackends, backend)
			}
			continue
		}
	}

	return missingBackends, nil
}

// serviceHasDomain validates whether the service has a domain defined.
func serviceHasDomain(sid string, sv int, apiClient api.Interface) (bool, error) {
	domains, err := apiClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      sid,
		ServiceVersion: sv,
	})
	if err != nil {
		return false, fmt.Errorf("error fetching service domains: %w", err)
	}
	if len(domains) < 1 {
		return false, nil
	}
	return true, nil
}

// configureDomain configures the domain value.
func configureDomain(c *DeployCommand, def string, out io.Writer, in io.Reader, f validator) (string, error) {
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

// validator represents a function that validates an input.
type validator func(input string) error

// validateDomain ensures the input domain looks like a domain.
func validateDomain(input string) error {
	if input == "" {
		return nil
	}
	if !domainNameRegEx.MatchString(input) {
		return fmt.Errorf("must be valid domain name")
	}
	return nil
}

// configureBackends determines whether multiple backends need to be configured
// or a singular backend based on existence of the fastly.toml [setup] table.
func configureBackends(c *DeployCommand, backendsToCreate []Backend, missingBackends []manifest.Mapper, out io.Writer, in io.Reader) ([]Backend, error) {
	if len(missingBackends) > 0 {
		for i, backend := range missingBackends {
			b, err := configurePredefinedBackend(i, backend, c, out, in, validateBackend)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
					"Backends (predefined)": missingBackends,
				})
				return nil, err
			}
			backendsToCreate = append(backendsToCreate, b)
		}
	} else {
		backends, err := configurePromptBackends(c, out, in, validateBackend)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Accept defaults": c.AcceptDefaults,
			})
			return nil, err
		}
		backendsToCreate = append(backendsToCreate, backends...)
	}
	return backendsToCreate, nil
}

// configurePredefinedBackend configures the backend address and its port number
// using values provided by the fastly.toml [setup] configuration.
func configurePredefinedBackend(i int, backend manifest.Mapper, c *DeployCommand, out io.Writer, in io.Reader, v validator) (Backend, error) {
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

	innerErr := fmt.Errorf("error parsing the [[setup.backends]] configuration")
	remediation := "Check the fastly.toml configuration for a missing or invalid '%s' field."

	// ADDRESS (REQUIRED)
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

	// NAME
	{
		if _, ok = backend["name"]; ok {
			name, ok = backend["name"].(string)
			if !ok {
				return b, backendRemediationError("name", remediation, innerErr)
			}
		}
		b.Name = name
		if name == "" {
			i = i + 1
			b.Name = fmt.Sprintf("backend_%d", i)
		}
	}

	// PROMPT
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

	// PORT
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

// configurePromptBackends configures multiple backends based on prompted input
// from the user.
//
// NOTE: If `--accept-defaults` is set, then create a single "originless" backend.
func configurePromptBackends(c *DeployCommand, out io.Writer, in io.Reader, f validator) (backends []Backend, err error) {
	if c.AcceptDefaults {
		backend := createOriginlessBackend()
		backends = append(backends, backend)
		return backends, nil
	}

	var i int
	for {
		var backend Backend

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

		backend.Name, err = text.Input(out, "Backend name: ", in)
		if err != nil {
			return backends, fmt.Errorf("error reading input %w", err)
		}
		if backend.Name == "" {
			i = i + 1
			backend.Name = fmt.Sprintf("backend_%d", i)
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

// validateBackend ensures the input backend is a valid hostname or IP.
//
// NOTE: An empty value is allowed because it allows the caller to
// short-circuit logic related to whether the user is prompted endlessly.
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

// createDomain creates the given domain and handles unrolling the stack in case
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

// createBackend creates the given backend and handles unrolling the stack in case
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
		}
		return fmt.Errorf("error creating backend: %w", err)
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

// getHashSum creates a SHA 512 hash from the given path input.
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
