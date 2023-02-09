package compute

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/setup"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/undo"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
)

const (
	manageServiceBaseURL = "https://manage.fastly.com/configure/services/"
	trialNotActivated    = "Valid values for 'type' are: 'vcl'"
)

// PackageSizeLimit describes the package size limit in bytes (currently 50mb)
// https://docs.fastly.com/products/compute-at-edge-billing-and-resource-limits#resource-limits
var PackageSizeLimit int64 = 50000000

// DeployCommand deploys an artifact previously produced by build.
type DeployCommand struct {
	cmd.Base

	// NOTE: these are public so that the "publish" composite command can set the
	// values appropriately before calling the Exec() function.
	Comment        cmd.OptionalString
	Domain         string
	Manifest       manifest.Data
	Package        string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = g
	c.Manifest = m
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute@Edge service")

	// NOTE: when updating these flags, be sure to update the composite command:
	// `compute publish`.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceVersion.Set,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Name:        cmd.FlagVersionName,
	})
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.Domain)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	fnActivateTrial, source, serviceID, pkgPath, hashSum, err := setupDeploy(c, out)
	if err != nil {
		return err
	}

	undoStack := undo.NewStack()
	undoStack.Push(func() error {
		// We'll only clean-up the service if it's a new service.
		//
		// SourceUndefined means...
		//   - the flags --service-id/--service-name were not set.
		//   - no service_id attribute was set in the fastly.toml manifest.
		if source == manifest.SourceUndefined {
			return cleanupService(c.Globals.APIClient, serviceID, c.Manifest, out)
		}
		return nil
	})

	newService, serviceID, serviceVersion, cont, err := serviceManagement(serviceID, source, c, in, out, fnActivateTrial)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	domains, backends, dictionaries, loggers, objectStores, err := constructSetupObjects(
		newService, serviceID, serviceVersion.Number, c, in, out,
	)
	if err != nil {
		return err
	}

	if err := processSetupConfig(
		newService, domains, backends, dictionaries, loggers, objectStores,
		serviceID, serviceVersion.Number, c, out,
	); err != nil {
		return err
	}

	progress := text.ResetProgress(out, c.Globals.Verbose())

	defer func(errLog fsterr.LogInterface, progress text.Progress) {
		if err != nil {
			errLog.Add(err)
			progress.Fail()
		}
		undoStack.RunIfError(out, err)
	}(c.Globals.ErrLog, progress)

	if err := processSetupCreation(
		newService, domains, backends, dictionaries, objectStores, progress, c,
		serviceID, serviceVersion.Number, out,
	); err != nil {
		return err
	}

	cont, err = processPackage(
		c, hashSum, pkgPath, serviceID, serviceVersion.Number, progress, out,
	)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	if err := processService(c, serviceID, serviceVersion.Number, progress); err != nil {
		return err
	}

	progress.Done()
	text.Break(out)
	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))

	displayDomain(c.Globals.APIClient, serviceID, serviceVersion.Number, out)

	text.Success(out, "Deployed package (service %s, version %v)", serviceID, serviceVersion.Number)
	return nil
}

// setupDeploy prepares the environment.
// It will do things like:
//   - Check if there is an API token missing.
//   - Acquire the Service ID/Version.
//   - Validate there is a package to deploy.
//   - Determine if a trial needs to be activated on the user's account.
func setupDeploy(c *DeployCommand, out io.Writer) (
	fnActivateTrial activator,
	source manifest.Source,
	serviceID, pkgPath, hashSum string,
	err error,
) {
	defaultActivator := func(customerID string) error { return nil }

	token, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return defaultActivator, 0, "", "", "", fsterr.ErrNoToken
	}

	// IMPORTANT: We don't handle the error when looking up the Service ID.
	// This is because later in the Exec() flow we might create a 'new' service.
	// Refer to manageNoServiceIDFlow()
	serviceID, source, flag, err := cmd.ServiceID(c.ServiceName, c.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err == nil && c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	pkgPath, hashSum, err = validatePackage(c.Manifest, c.Package, c.Globals.Verbose(), c.Globals.ErrLog, out)
	if err != nil {
		return defaultActivator, source, serviceID, "", "", err
	}

	endpoint, _ := c.Globals.Endpoint()
	fnActivateTrial = preconfigureActivateTrial(endpoint, token, c.Globals.HTTPClient)

	return fnActivateTrial, source, serviceID, pkgPath, hashSum, err
}

// validatePackage short-circuits the deploy command if the user hasn't first
// built a package to be deployed.
//
// NOTE: It also validates if the package size exceeds limit:
// https://docs.fastly.com/products/compute-at-edge-billing-and-resource-limits#resource-limits
func validatePackage(
	data manifest.Data,
	packageFlag string,
	verbose bool,
	errLog fsterr.LogInterface,
	out io.Writer,
) (pkgPath, hashSum string, err error) {
	err = data.File.ReadError()
	if err != nil {
		if packageFlag == "" {
			if errors.Is(err, os.ErrNotExist) {
				err = fsterr.ErrReadingManifest
			}
			return pkgPath, hashSum, err
		}

		// NOTE: Before returning the manifest read error, we'll attempt to read
		// the manifest from within the given package archive.
		err := readManifestFromPackageArchive(&data, packageFlag, verbose, out)
		if err != nil {
			return pkgPath, hashSum, err
		}
	}

	projectName, source := data.Name()
	pkgPath, err = packagePath(packageFlag, projectName, source)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": packageFlag,
		})
		return pkgPath, hashSum, err
	}

	pkgSize, err := packageSize(pkgPath)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": pkgPath,
		})
		return pkgPath, hashSum, fsterr.RemediationError{
			Inner:       fmt.Errorf("error reading package size: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}

	if pkgSize > PackageSizeLimit {
		return pkgPath, hashSum, fsterr.RemediationError{
			Inner:       fmt.Errorf("package size is too large (%d bytes)", pkgSize),
			Remediation: fsterr.PackageSizeRemediation,
		}
	}

	contents := map[string]*bytes.Buffer{
		"fastly.toml": {},
		"main.wasm":   {},
	}
	if err := validate(pkgPath, func(f archiver.File) error {
		switch fname := f.Name(); fname {
		case "fastly.toml", "main.wasm":
			if _, err := io.Copy(contents[fname], f); err != nil {
				return fmt.Errorf("error reading %s: %w", fname, err)
			}
		}
		return nil
	}); err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": pkgPath,
			"Package size": pkgSize,
		})
		return pkgPath, hashSum, err
	}

	hashSum, err = getHashSum(contents)
	if err != nil {
		return pkgPath, "", err
	}

	return pkgPath, hashSum, nil
}

// readManifestFromPackageArchive extracts the manifest file from the given
// package archive file and reads it into memory.
func readManifestFromPackageArchive(data *manifest.Data, packageFlag string, verbose bool, out io.Writer) error {
	dst, err := os.MkdirTemp("", fmt.Sprintf("%s-*", manifest.Filename))
	if err != nil {
		return err
	}
	defer os.RemoveAll(dst)

	if err = archiver.Unarchive(packageFlag, dst); err != nil {
		return fmt.Errorf("error extracting package '%s': %w", packageFlag, err)
	}

	files, err := os.ReadDir(dst)
	if err != nil {
		return err
	}
	extractedDirName := files[0].Name()

	manifestPath, err := locateManifest(filepath.Join(dst, extractedDirName))
	if err != nil {
		return err
	}

	err = data.File.Read(manifestPath)
	if err != nil {
		return err
	}

	if verbose {
		text.Info(out, "Using fastly.toml within --package archive:\n\t%s", packageFlag)
	}

	return nil
}

// locateManifest attempts to find the manifest within the given path's
// directory tree.
func locateManifest(path string) (string, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var foundManifest string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Base(path) == manifest.Filename {
			foundManifest = path
			return fsterr.ErrStopWalk
		}
		return nil
	})

	if err != nil {
		// If the error isn't ErrStopWalk, then the WalkDir() function had an
		// issue processing the directory tree.
		if err != fsterr.ErrStopWalk {
			return "", err
		}

		return foundManifest, nil
	}

	return "", fmt.Errorf("error locating manifest within the given path: %s", path)
}

// packagePath generates a path that points to a package tar inside the pkg
// directory if the package path flag was not set by the user.
func packagePath(path, projectName string, source manifest.Source) (string, error) {
	if path == "" {
		if source == manifest.SourceUndefined {
			return "", fsterr.ErrReadingManifest
		}
		path = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
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

// activator represents a function that calls an undocumented API endpoint for
// activating a Compute@Edge free trial on the given customer account.
//
// It is preconfigured with the Fastly API endpoint, a user token and a simple
// HTTP Client.
//
// This design allows us to pass an activator rather than passing multiple
// unrelated arguments through several nested functions.
type activator func(customerID string) error

// preconfigureActivateTrial activates a free trial on the customer account.
func preconfigureActivateTrial(endpoint, token string, httpClient api.HTTPClient) activator {
	return func(customerID string) error {
		path := fmt.Sprintf(undocumented.EdgeComputeTrial, customerID)
		_, err := undocumented.Get(endpoint, path, token, httpClient)
		if err != nil {
			apiErr, ok := err.(undocumented.APIError)
			if !ok {
				return err
			}
			// 409 Conflict == The Compute@Edge trial has already been created.
			if apiErr.StatusCode != http.StatusConflict {
				return fmt.Errorf("%w: %d %s", err, apiErr.StatusCode, http.StatusText(apiErr.StatusCode))
			}
		}
		return nil
	}
}

func serviceManagement(
	serviceID string,
	source manifest.Source,
	c *DeployCommand,
	in io.Reader,
	out io.Writer,
	fnActivateTrial activator,
) (newService bool, updatedServiceID string, serviceVersion *fastly.Version, cont bool, err error) {
	if source == manifest.SourceUndefined {
		newService = true
		serviceID, serviceVersion, err = manageNoServiceIDFlow(c.Globals.Flags, in, out, c.Globals.Verbose(), c.Globals.APIClient, c.Package, c.Globals.ErrLog, &c.Manifest.File, fnActivateTrial)
		if err != nil {
			return newService, "", nil, false, err
		}
		if serviceID == "" {
			return newService, "", nil, false, nil // user declined service creation prompt
		}
	} else {
		serviceVersion, err = manageExistingServiceFlow(serviceID, c.ServiceVersion, c.Globals.APIClient, c.Globals.Verbose(), out, c.Globals.ErrLog)
		if err != nil {
			return false, "", nil, false, err
		}
	}

	return newService, serviceID, serviceVersion, true, nil
}

// manageNoServiceIDFlow handles creating a new service when no Service ID is found.
func manageNoServiceIDFlow(
	f global.Flags,
	in io.Reader,
	out io.Writer,
	verbose bool,
	apiClient api.Interface,
	packageFlag string,
	errLog fsterr.LogInterface,
	manifestFile *manifest.File,
	fnActivateTrial activator,
) (serviceID string, serviceVersion *fastly.Version, err error) {
	if !f.AutoYes && !f.NonInteractive {
		text.Break(out)
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")
		text.Break(out)

		answer, err := text.AskYesNo(out, text.BoldYellow("Create new service: [y/N] "), in)
		if err != nil {
			return serviceID, serviceVersion, err
		}
		if !answer {
			return serviceID, serviceVersion, nil
		}

		text.Break(out)
	}

	defaultServiceName := manifestFile.Name
	var serviceName string

	if !f.AcceptDefaults && !f.NonInteractive {
		serviceName, err = text.Input(out, text.BoldYellow(fmt.Sprintf("Service name: [%s] ", defaultServiceName)), in)
		if err != nil || serviceName == "" {
			serviceName = defaultServiceName
		}
	} else {
		serviceName = defaultServiceName
	}

	progress := text.NewProgress(out, verbose)

	// There is no service and so we'll do a one time creation of the service
	//
	// NOTE: we're shadowing the `serviceVersion` and `serviceID` variables.
	serviceID, serviceVersion, err = createService(serviceName, apiClient, fnActivateTrial, progress, errLog)
	if err != nil {
		progress.Fail()
		errLog.AddWithContext(err, map[string]any{
			"Service name": serviceName,
		})
		return serviceID, serviceVersion, err
	}

	progress.Done()

	// NOTE: Only attempt to update the manifest if the user has not specified
	// the --package flag, as this suggests they are not inside a project
	// directory and subsequently we're reading the manifest content from within
	// a given .tar.gz package archive file.
	if packageFlag == "" {
		err = updateManifestServiceID(manifestFile, manifest.Filename, serviceID)
		if err != nil {
			errLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return serviceID, serviceVersion, err
		}
	}

	text.Break(out)
	return serviceID, serviceVersion, nil
}

// createService creates a service to associate with the compute package.
//
// NOTE: If the creation of the service fails because the user has not
// activated a free trial, then we'll trigger the trial for their account.
func createService(
	serviceName string,
	apiClient api.Interface,
	fnActivateTrial activator,
	progress text.Progress,
	errLog fsterr.LogInterface,
) (serviceID string, serviceVersion *fastly.Version, err error) {
	progress.Step("Creating service...")

	service, err := apiClient.CreateService(&fastly.CreateServiceInput{
		Name: &serviceName,
		Type: fastly.String("wasm"),
	})
	if err != nil {
		if strings.Contains(err.Error(), trialNotActivated) {
			user, err := apiClient.GetCurrentUser()
			if err != nil {
				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       fmt.Errorf("unable to identify user associated with the given token: %w", err),
					Remediation: "To ensure you have access to the Compute@Edge platform we need your Customer ID. " + fsterr.AuthRemediation,
				}
			}

			err = fnActivateTrial(user.CustomerID)
			if err != nil {
				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       fmt.Errorf("error creating service: you do not have the Compute@Edge free trial enabled on your Fastly account"),
					Remediation: fsterr.ComputeTrialRemediation,
				}
			}

			errLog.AddWithContext(err, map[string]any{
				"Service Name": serviceName,
				"Customer ID":  user.CustomerID,
			})
			return createService(serviceName, apiClient, fnActivateTrial, progress, errLog)
		}

		errLog.AddWithContext(err, map[string]any{
			"Service Name": serviceName,
		})
		return serviceID, serviceVersion, fmt.Errorf("error creating service: %w", err)
	}

	return service.ID, &fastly.Version{Number: 1}, nil
}

// cleanupService is executed if a new service flow has errors.
// It deletes the service, which will cause any contained resources to be deleted.
// It will also strip the Service ID from the fastly.toml manifest file.
func cleanupService(apiClient api.Interface, serviceID string, m manifest.Data, out io.Writer) error {
	text.Info(out, "Cleaning up service")

	err := apiClient.DeleteService(&fastly.DeleteServiceInput{
		ID: serviceID,
	})
	if err != nil {
		return err
	}

	text.Info(out, "Removing Service ID from fastly.toml")

	err = updateManifestServiceID(&m.File, manifest.Filename, "")
	if err != nil {
		return err
	}

	text.Output(out, "Cleanup complete")
	return nil
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
	errLog fsterr.LogInterface,
) (serviceVersion *fastly.Version, err error) {
	serviceVersion, err = serviceVersionFlag.Parse(serviceID, apiClient)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return serviceVersion, err
	}

	// Validate that we're dealing with a Compute@Edge 'wasm' service and not a
	// VCL service, for which we cannot upload a wasm package format to.
	serviceDetails, err := apiClient.GetServiceDetails(&fastly.GetServiceInput{ID: serviceID})
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return serviceVersion, err
	}
	if serviceDetails.Type != "wasm" {
		errLog.AddWithContext(fmt.Errorf("error: invalid service type: '%s'", serviceDetails.Type), map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
			"Service Type":    serviceDetails.Type,
		})
		return serviceVersion, fsterr.RemediationError{
			Inner:       fmt.Errorf("invalid service type: %s", serviceDetails.Type),
			Remediation: "Ensure the provided Service ID is associated with a 'Wasm' Fastly Service and not a 'VCL' Fastly service. " + fsterr.ComputeTrialRemediation,
		}
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
func errLogService(l fsterr.LogInterface, err error, sid string, sv int) {
	l.AddWithContext(err, map[string]any{
		"Service ID":      sid,
		"Service Version": sv,
	})
}

// checkServiceID validates the given Service ID maps to a real service.
func checkServiceID(serviceID string, client api.Interface) error {
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
func pkgCompare(client api.Interface, serviceID string, version int, hashSum string, progress text.Progress, out io.Writer) (bool, error) {
	p, err := client.GetPackage(&fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})

	if err == nil {
		if hashSum == p.Metadata.HashSum {
			progress.Done()
			text.Info(out, "Skipping package deployment, local and service version are identical. (service %v, version %v) ", serviceID, version)
			return false, nil
		}
	}

	return true, nil
}

// getHashSum creates a SHA 512 hash from the given file contents in a specific order.
func getHashSum(contents map[string]*bytes.Buffer) (hash string, err error) {
	h := sha512.New()
	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, fname := range keys {
		if _, err := io.Copy(h, contents[fname]); err != nil {
			return "", err
		}
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

// displayDomain displays a domain from those available in the service.
func displayDomain(apiClient api.Interface, serviceID string, serviceVersion int, out io.Writer) {
	latestDomains, err := apiClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	})
	if err == nil {
		name := latestDomains[0].Name
		if segs := strings.Split(name, "*."); len(segs) > 1 {
			name = segs[1]
		}
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", name))
	}
}

func constructSetupObjects(
	newService bool,
	serviceID string,
	serviceVersion int,
	c *DeployCommand,
	in io.Reader,
	out io.Writer,
) (
	*setup.Domains,
	*setup.Backends,
	*setup.Dictionaries,
	*setup.Loggers,
	*setup.ObjectStores,
	error,
) {
	var err error

	// We only check the Service ID is valid when handling an existing service.
	if !newService {
		err = checkServiceID(serviceID, c.Globals.APIClient)
		if err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return nil, nil, nil, nil, nil, err
		}
	}

	// Because a service_id exists in the fastly.toml doesn't mean it's valid
	// e.g. it could be missing required resources such as a domain or backend.
	// We check and allow the user to configure these settings before continuing.

	domains := &setup.Domains{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		PackageDomain:  c.Domain,
		RetryLimit:     5,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Stdin:          in,
		Stdout:         out,
		Verbose:        c.Globals.Verbose(),
	}

	err = domains.Validate()
	if err != nil {
		errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
		return nil, nil, nil, nil, nil, fmt.Errorf("error configuring service domains: %w", err)
	}

	var (
		backends     *setup.Backends
		dictionaries *setup.Dictionaries
		loggers      *setup.Loggers
		objectStores *setup.ObjectStores
	)

	if newService {
		backends = &setup.Backends{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.Backends,
			Stdin:          in,
			Stdout:         out,
		}

		dictionaries = &setup.Dictionaries{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.Dictionaries,
			Stdin:          in,
			Stdout:         out,
		}

		loggers = &setup.Loggers{
			Setup:  c.Manifest.File.Setup.Loggers,
			Stdout: out,
		}

		objectStores = &setup.ObjectStores{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.ObjectStores,
			Stdin:          in,
			Stdout:         out,
		}
	}

	return domains, backends, dictionaries, loggers, objectStores, nil
}

func processSetupConfig(
	newService bool,
	domains *setup.Domains,
	backends *setup.Backends,
	dictionaries *setup.Dictionaries,
	loggers *setup.Loggers,
	objectStores *setup.ObjectStores,
	serviceID string,
	serviceVersion int,
	c *DeployCommand,
	out io.Writer,
) (err error) {
	if domains.Missing() {
		err = domains.Configure()
		if err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service domains: %w", err)
		}
	}

	// IMPORTANT: The pointer refs in this block are not checked for nil.
	// We presume if we're dealing with newService they have been set.
	if newService {
		// NOTE: A service can't be activated without at least one backend defined.
		// This explains why the following block of code isn't wrapped in a call to
		// the .Predefined() method, as the call to .Configure() will ensure the
		// user is prompted regardless of whether there is a [setup.backends]
		// defined in the fastly.toml configuration.
		err = backends.Configure()
		if err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service backends: %w", err)
		}

		if dictionaries.Predefined() {
			err = dictionaries.Configure()
			if err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service dictionaries: %w", err)
			}
		}

		if loggers.Predefined() {
			// NOTE: We don't handle errors from the Configure() method because we
			// don't actually do anything other than display a message to the user
			// informing them that they need to create a log endpoint and which
			// provider type they should be. The reason we don't implement logic for
			// creating logging objects is because the API input fields vary
			// significantly between providers.
			_ = loggers.Configure()
		}

		if objectStores.Predefined() {
			err = objectStores.Configure()
			if err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service object stores: %w", err)
			}
		}
	}

	text.Break(out)

	return nil
}

func processSetupCreation(
	newService bool,
	domains *setup.Domains,
	backends *setup.Backends,
	dictionaries *setup.Dictionaries,
	objectStores *setup.ObjectStores,
	progress text.Progress,
	c *DeployCommand,
	serviceID string,
	serviceVersion int,
	out io.Writer,
) error {
	// NOTE: We need to output this message immediately to avoid breaking prompt.
	if newService {
		text.Info(out, "Processing of the fastly.toml [setup] configuration happens only when there is no existing service. Once a service is created, any further changes to the service or its resources must be made manually.")
		text.Break(out)
	}

	if domains.Missing() {
		// NOTE: We can't pass a text.Progress instance to setup.Domains at the
		// point of constructing the domains object, as the text.Progress instance
		// prevents other stdout from being read.
		domains.Progress = progress

		if err := domains.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}
	}

	// IMPORTANT: The pointer refs in this block are not checked for nil.
	// We presume if we're dealing with newService they have been set.
	if newService {
		// NOTE: We can't pass a text.Progress instance to setup.Backends or
		// setup.Dictionaries (etc) at the point of constructing the setup objects,
		// as the text.Progress instance prevents other stdout from being read.
		backends.Progress = progress
		dictionaries.Progress = progress
		objectStores.Progress = progress

		if err := backends.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := dictionaries.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := objectStores.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}
	}

	return nil
}

func processPackage(
	c *DeployCommand,
	hashSum, pkgPath, serviceID string,
	serviceVersion int,
	progress text.Progress,
	out io.Writer,
) (cont bool, err error) {
	cont, err = pkgCompare(c.Globals.APIClient, serviceID, serviceVersion, hashSum, progress, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path":    pkgPath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return false, err
	}
	if !cont {
		return false, nil
	}

	err = pkgUpload(progress, c.Globals.APIClient, serviceID, serviceVersion, pkgPath)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path":    pkgPath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return false, err
	}

	return true, nil
}

func processService(c *DeployCommand, serviceID string, serviceVersion int, progress text.Progress) error {
	if c.Comment.WasSet {
		_, err := c.Globals.APIClient.UpdateVersion(&fastly.UpdateVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Comment:        &c.Comment.Value,
		})
		if err != nil {
			return fmt.Errorf("error setting comment for service version %d: %w", serviceVersion, err)
		}
	}

	progress.Step("Activating version...")

	_, err := c.Globals.APIClient.ActivateVersion(&fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return fmt.Errorf("error activating version: %w", err)
	}

	return nil
}
