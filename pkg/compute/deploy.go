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
	"github.com/fastly/cli/pkg/compute/manifest"
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
	Manifest       manifest.Data
	Path           string
	Domain         string
	Backend        string
	BackendPort    uint
	ServiceVersion cmd.OptionalServiceVersion
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
	c.CmdClause.Flag("path", "Path to package").Short('p').StringVar(&c.Path)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.Domain)
	c.CmdClause.Flag("backend", "A hostname, IPv4, or IPv6 address for the package backend").StringVar(&c.Backend)
	c.CmdClause.Flag("backend-port", "A port number for the package backend").UintVar(&c.BackendPort)
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
		c.Globals.ErrLog.Add(err)
		return err
	}
	if err := validate(path); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	var (
		domain, backend string
		backendPort     uint
		invalidService  bool
		invalidType     invalidResource
		version         *fastly.Version
	)

	serviceID, sidSrc := c.Manifest.ServiceID()
	if sidSrc == manifest.SourceUndefined {
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")
		text.Break(out)

		domain, err = cfgDomain(c.Domain, defaultTopLevelDomain, out, in, validateDomain)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Break(out)
	} else {
		version, err = c.ServiceVersion.Parse(serviceID, c.Globals.Client)
		if err != nil {
			c.Globals.ErrLog.Add(err)
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
				c.Globals.ErrLog.Add(err)
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
		ok, invalidType, err = validateService(serviceID, c.Globals.Client, version)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if !ok {
			invalidService = true

			text.Output(out, "Service '%s' is missing required domain or backend. These must be added before the Compute@Edge service can be deployed.", serviceID)
			text.Break(out)

			switch invalidType {
			case resourceBoth:
				domain, err = cfgDomain(c.Domain, defaultTopLevelDomain, out, in, validateDomain)
				if err != nil {
					c.Globals.ErrLog.Add(err)
					return err
				}
				backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
				if err != nil {
					c.Globals.ErrLog.Add(err)
					return err
				}
			case resourceDomain:
				domain, err = cfgDomain(c.Domain, defaultTopLevelDomain, out, in, validateDomain)
				if err != nil {
					c.Globals.ErrLog.Add(err)
					return err
				}
			case resourceBackend:
				backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
				if err != nil {
					c.Globals.ErrLog.Add(err)
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
		// and the associated domain/backend and store the Service ID within the
		// manifest. On subsequent runs of the deploy subcommand we'll skip the
		// service/domain/backend creation.
		//
		// NOTE: we're shadowing the `version` and `serviceID` variable.
		version, serviceID, err = createService(progress, c.Globals.Client, name, desc)
		if err != nil {
			c.Globals.ErrLog.Add(err)
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
			c.Globals.ErrLog.Add(err)
			return err
		}
		err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		err = updateManifestServiceID(&c.Manifest.File, manifest.Filename, progress, serviceID)
		if err != nil {
			c.Globals.ErrLog.Add(err)
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
				c.Globals.ErrLog.Add(err)
				return err
			}
			err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		case resourceDomain:
			err = createDomain(progress, c.Globals.Client, serviceID, version.Number, domain, undoStack)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		case resourceBackend:
			err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		}
	}

	cont, err := pkgCompare(c.Globals.Client, serviceID, version.Number, path, progress, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if !cont {
		return nil
	}

	err = pkgUpload(progress, c.Globals.Client, serviceID, version.Number, path)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	progress.Step("Activating version...")

	_, err = c.Globals.Client.ActivateVersion(&fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
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
// defined.
//
// The use of `resourceNone` as a `invalidResource` enum type is to represent a
// situation where an error occurred and so we were unable to identify whether
// a domain or backend (or both) were missing from the given service.
func validateService(serviceID string, client api.Interface, version *fastly.Version) (bool, invalidResource, error) {
	err := checkServiceID(serviceID, client, version)
	if err != nil {
		return false, resourceNone, err
	}

	domains, err := client.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		return false, resourceNone, fmt.Errorf("error fetching service domains: %w", err)
	}

	backends, err := client.ListBackends(&fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
		return false, resourceNone, fmt.Errorf("error fetching service backends: %w", err)
	}

	ld := len(domains)
	lb := len(backends)

	if ld == 0 && lb == 0 {
		return false, resourceBoth, nil
	}
	if ld == 0 {
		return false, resourceDomain, nil
	}
	if lb == 0 {
		return false, resourceBackend, nil
	}

	return true, resourceNone, nil
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
func cfgDomain(domain string, def string, out io.Writer, in io.Reader, f validator) (string, error) {
	if domain != "" {
		return domain, nil
	}

	rand.Seed(time.Now().UnixNano())

	defaultDomain := fmt.Sprintf("%s.%s", petname.Generate(3, "-"), def)
	domain, err := text.Input(out, fmt.Sprintf("Domain: [%s] ", defaultDomain), in, f)
	if err != nil {
		return "", fmt.Errorf("error reading input %w", err)
	}

	if domain == "" {
		return defaultDomain, nil
	}

	return domain, nil
}

// cfgBackend configures the backend address and its port number values.
func cfgBackend(backend string, backendPort uint, out io.Writer, in io.Reader, f validator) (string, uint, error) {
	if backend == "" {
		var err error
		backend, err = text.Input(out, "Backend (originless, hostname or IP address): [originless] ", in, f)

		if err != nil {
			return "", 0, fmt.Errorf("error reading input %w", err)
		}

		if backend == "" || backend == "originless" {
			backend = "127.0.0.1"
			backendPort = uint(80)
		}
	}

	if backendPort == 0 {
		input, err := text.Input(out, "Backend port number: [80] ", in)
		if err != nil {
			return "", 0, fmt.Errorf("error reading input %w", err)
		}

		portnumber := 80
		if input != "" {
			portnumber, err = strconv.Atoi(input)
			if err != nil {
				text.Warning(out, "error converting input. We'll use the default port number: [80].")
				portnumber = 80
			}
		}

		backendPort = uint(portnumber)
	}

	return backend, backendPort, nil
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
func createBackend(progress text.Progress, client api.Interface, serviceID string, version int, backend string, backendPort uint, undoStack undo.Stacker) error {
	progress.Step("Creating backend...")

	undoStack.Push(func() error {
		return client.DeleteBackend(&fastly.DeleteBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           backend,
		})
	})

	_, err := client.CreateBackend(&fastly.CreateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           backend,
		Address:        backend,
		Port:           backendPort,
	})
	if err != nil {
		return fmt.Errorf("error creating backend: %w", err)
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
		isHost = true
	}
	isEmpty := input == ""
	isOriginless := strings.ToLower(input) == "originless"
	if !isEmpty && !isOriginless && !isHost && !isAddr {
		return fmt.Errorf(`must be "originless" or a valid hostname, IPv4, or IPv6 address`)
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
