package compute

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/commands/compute/setup"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/undo"
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/kennygrant/sanitize"
)

const (
	manageServiceBaseURL = "https://manage.fastly.com/configure/services/"
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
	Comment        cmd.OptionalString
	Domain         string
	Manifest       manifest.Data
	Path           string
	ServiceVersion cmd.OptionalServiceVersion
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

	domains := &setup.Domains{
		AcceptDefaults: c.AcceptDefaults,
		APIClient:      apiClient,
		PackageDomain:  c.Domain,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
		Stdin:          in,
		Stdout:         out,
	}

	err = domains.Validate()
	if err != nil {
		errLogService(errLog, err, serviceID, serviceVersion.Number)
		return fmt.Errorf("error configuring service domains: %w", err)
	}

	var backends *setup.Backends

	if newService {
		backends = &setup.Backends{
			AcceptDefaults: c.AcceptDefaults,
			APIClient:      apiClient,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion.Number,
			Setup:          c.Manifest.File.Setup.Backends,
			Stdin:          in,
			Stdout:         out,
		}
	}

	// RESOURCE CONFIGURATION...

	if domains.Missing() {
		err = domains.Configure()
		if err != nil {
			errLogService(errLog, err, serviceID, serviceVersion.Number)
			return fmt.Errorf("error configuring service domains: %w", err)
		}
	}

	if newService {
		err = backends.Configure()
		if err != nil {
			errLogService(errLog, err, serviceID, serviceVersion.Number)
			return fmt.Errorf("error configuring service backends: %w", err)
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

	if domains.Missing() {
		// NOTE: We can't pass a text.Progress instance to setup.Domains at the
		// point of constructing the domains object, as the text.Progress instance
		// prevents other stdout from being read.
		domains.Progress = progress

		if err := domains.Create(); err != nil {
			errLog.AddWithContext(err, map[string]interface{}{
				"Accept defaults": c.AcceptDefaults,
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
	}

	if newService {
		// NOTE: We can't pass a text.Progress instance to setup.Backends at the
		// point of constructing the backends object, as the text.Progress instance
		// prevents other stdout from being read.
		backends.Progress = progress

		if err := backends.Create(); err != nil {
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

	latestDomains, err := apiClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	})
	if err == nil {
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", latestDomains[0].Name))
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
	err = data.File.Read(manifest.Filename)
	if err != nil {
		return pkgName, pkgPath, err
	}
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
