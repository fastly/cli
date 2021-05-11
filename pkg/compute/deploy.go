package compute

import (
	"crypto/sha512"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
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
	common.Base
	manifest manifest.Data

	// NOTE: these are public so that the "publish" composite command can set the
	// values appropriately before calling the Exec() function.
	Path        string
	Version     common.OptionalInt
	Domain      string
	Backend     string
	BackendPort uint
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute@Edge service")

	// NOTE: when updating these flags, be sure to update the composite command:
	// `compute publish`.
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version to activate").Action(c.Version.Set).IntVar(&c.Version.Value)
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

	var (
		domain, backend string
		backendPort     uint
		invalidService  bool
		invalidType     invalidResource
		version         *fastly.Version
	)

	serviceID, sidSrc := c.manifest.ServiceID()
	if sidSrc == manifest.SourceUndefined {
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")
		text.Break(out)

		domain, err = cfgDomain(c.Domain, defaultTopLevelDomain, out, in, validateDomain)
		if err != nil {
			return err
		}

		backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
		if err != nil {
			return err
		}

		text.Break(out)
	} else {
		// We define the `ok` variable so that the following call to
		// `validateservice` will be able to shadow multiple variables.
		var ok bool

		// Because a service_id exists in the fastly.toml doesn't mean it's valid
		// e.g. it could be missing either a domain or backend resource. So we
		// check and allow the user to configure these settings before continuing.
		//
		// The returned version will only be nil if the initial service version
		// request failed.
		version, ok, invalidType, err = validateService(serviceID, c.Globals.Client, c.Version)
		if err != nil {
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
					return err
				}
				backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
				if err != nil {
					return err
				}
			case resourceDomain:
				domain, err = cfgDomain(c.Domain, defaultTopLevelDomain, out, in, validateDomain)
				if err != nil {
					return err
				}
			case resourceBackend:
				backend, backendPort, err = cfgBackend(c.Backend, c.BackendPort, out, in, validateBackend)
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

	undoStack := common.NewUndoStack()

	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
		undoStack.RunIfError(out, err)
	}()

	name, source := c.manifest.Name()
	path, err := pkgPath(c.Path, progress, name, source)
	if err != nil {
		return err
	}

	if sidSrc == manifest.SourceUndefined {
		// There is no service and so we'll do a one time creation of the service
		// and the associated domain/backend and store the Service ID within the
		// manifest. On subsequent runs of the deploy subcommand we'll skip the
		// service/domain/backend creation.
		//
		// NOTE: we're shadowing the `version` and `serviceID` variable.
		version, serviceID, err = createService(progress, c.Globals.Client, name, desc)
		if err != nil {
			return err
		}

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
			return err
		}
		err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
		if err != nil {
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
				return err
			}
			err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
			if err != nil {
				return err
			}
		case resourceDomain:
			err = createDomain(progress, c.Globals.Client, serviceID, version.Number, domain, undoStack)
			if err != nil {
				return err
			}
		case resourceBackend:
			err = createBackend(progress, c.Globals.Client, serviceID, version.Number, backend, backendPort, undoStack)
			if err != nil {
				return err
			}
		}
	}

	// We only want to update the manifest if it's the first time deploying and
	// so we will have only just created a new service. If the Service ID is
	// provided by the flag and not the file, then we'll also store that ID
	// within the manifest.
	if sidSrc == manifest.SourceUndefined || sidSrc != manifest.SourceFile {
		err = updateManifestServiceID(ManifestFilename, progress, serviceID)
		if err != nil {
			return err
		}
	}

	progress.Step("Validating package...")
	if err := validate(path); err != nil {
		return err
	}

	cont, err := pkgCompare(c.Globals.Client, serviceID, version.Number, path, progress, out)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	err = pkgUpload(progress, c.Globals.Client, serviceID, version.Number, path)
	if err != nil {
		return err
	}

	progress.Step("Activating version...")

	_, err = c.Globals.Client.ActivateVersion(&fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version.Number,
	})
	if err != nil {
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
func pkgPath(path string, progress text.Progress, name string, source manifest.Source) (string, error) {
	if path == "" {
		progress.Step("Reading package manifest...")

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
func validateService(serviceID string, client api.Interface, version common.OptionalInt) (*fastly.Version, bool, invalidResource, error) {
	v, err := serviceVersion(serviceID, client, version)
	if err != nil {
		return nil, false, resourceNone, err
	}

	domains, err := client.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: v.Number,
	})
	if err != nil {
		return v, false, resourceNone, fmt.Errorf("error fetching service domains: %w", err)
	}

	backends, err := client.ListBackends(&fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: v.Number,
	})
	if err != nil {
		return v, false, resourceNone, fmt.Errorf("error fetching service backends: %w", err)
	}

	ld := len(domains)
	lb := len(backends)

	if ld == 0 && lb == 0 {
		return v, false, resourceBoth, nil
	}
	if ld == 0 {
		return v, false, resourceDomain, nil
	}
	if lb == 0 {
		return v, false, resourceBackend, nil
	}

	return v, true, resourceNone, nil
}

// serviceVersion returns the version for the given service.
func serviceVersion(serviceID string, client api.Interface, versionFlag common.OptionalInt) (*fastly.Version, error) {
	_, err := client.GetService(&fastly.GetServiceInput{
		ID: serviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching service details: %w", err)
	}

	var version *fastly.Version

	if versionFlag.WasSet {
		version = &fastly.Version{Number: versionFlag.Value}
	} else {
		var err error
		version, err = pkgVersion(serviceID, client)
		if err != nil {
			return nil, err
		}
	}

	return version, nil
}

// pkgVersion acquires the ideal version to associate with a compute package.
func pkgVersion(serviceID string, client api.Interface) (*fastly.Version, error) {
	versions, err := client.ListVersions(&fastly.ListVersionsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing service versions: %w", err)
	}

	version, err := getLatestIdealVersion(versions)
	if err != nil {
		return nil, fmt.Errorf("error finding latest service version")
	}

	if version.Active || version.Locked {
		version, err := client.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: version.Number,
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning latest service version: %w", err)
		}

		return version, nil
	}

	return version, nil
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
				Remediation: "See more at https://fastly.dev/learning/compute/#create-a-new-fastly-account-and-invite-your-collaborators",
			}
		}
		return nil, "", fmt.Errorf("error creating service: %w", err)
	}

	version := &fastly.Version{Number: 1}
	serviceID := service.ID

	return version, serviceID, nil
}

// updateManifestServiceID updates the Service ID in the manifest.
func updateManifestServiceID(manifestFilename string, progress text.Progress, serviceID string) error {
	var m manifest.File

	if err := m.Read(manifestFilename); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	fmt.Fprintf(progress, "Setting service ID in manifest to %q...\n", serviceID)

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

		portnumber, err := strconv.Atoi(input)
		if err != nil {
			text.Warning(out, "error converting input: %v. We'll use the default port number: [80].", err)
			portnumber = 80
		}

		backendPort = uint(portnumber)
	}

	return backend, backendPort, nil
}

// createDomain creates the given domain and handle unrolling the stack in case
// of an error (i.e. will ensure the domain is deleted if there is an error).
func createDomain(progress text.Progress, client api.Interface, serviceID string, version int, domain string, undoStack common.Undoer) error {
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
func createBackend(progress text.Progress, client api.Interface, serviceID string, version int, backend string, backendPort uint, undoStack common.Undoer) error {
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

// getLatestIdealVersion gets the most ideal service version using the following logic:
// - Find the active version and return
// - If no active version, find the latest locked version and return
// - Otherwise return the latest version
func getLatestIdealVersion(versions []*fastly.Version) (*fastly.Version, error) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].UpdatedAt.Before(*versions[j].UpdatedAt)
	})

	var active, locked, latest *fastly.Version
	for i := 0; i < len(versions); i++ {
		v := versions[i]
		if v.Active {
			active = v
		}
		if v.Locked {
			locked = v
		}
		latest = v
	}

	var version *fastly.Version
	if active != nil {
		version = active
	} else if locked != nil {
		version = locked
	} else {
		version = latest
	}

	if version == nil {
		return nil, fmt.Errorf("error finding latest service version")
	}

	return version, nil
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
