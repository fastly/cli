package compute

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"

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
)

const (
	manageServiceBaseURL = "https://manage.fastly.com/configure/services/"
	trialNotActivated    = "Valid values for 'type' are: 'vcl'"
)

// DeployCommand deploys an artifact previously produced by build.
type DeployCommand struct {
	cmd.Base

	// NOTE: these are public so that the "publish" composite command can set the
	// values appropriately before calling the Exec() function.
	Comment            cmd.OptionalString
	Domain             string
	Manifest           manifest.Data
	Package            string
	ServiceName        cmd.OptionalServiceNameID
	ServiceVersion     cmd.OptionalServiceVersion
	StatusCheckCode    int
	StatusCheckOff     bool
	StatusCheckPath    string
	StatusCheckTimeout int
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
	c.CmdClause.Flag("status-check-code", "Set the expected status response for the service availability check").IntVar(&c.StatusCheckCode)
	c.CmdClause.Flag("status-check-off", "Disable the service availability check").BoolVar(&c.StatusCheckOff)
	c.CmdClause.Flag("status-check-path", "Specify the URL path for the service availability check").Default("/").StringVar(&c.StatusCheckPath)
	c.CmdClause.Flag("status-check-timeout", "Set a timeout (in seconds) for the service availability check").Default("120").IntVar(&c.StatusCheckTimeout)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	fnActivateTrial, source, serviceID, pkgPath, err := setupDeploy(c, out)
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

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	newService, serviceID, serviceVersion, cont, err := serviceManagement(serviceID, source, c, in, out, fnActivateTrial, spinner)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	so, err := constructSetupObjects(
		newService, serviceID, serviceVersion.Number, c, in, out,
	)
	if err != nil {
		return err
	}

	if err = processSetupConfig(
		newService, so, serviceID, serviceVersion.Number, c,
	); err != nil {
		return err
	}

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
		}
		undoStack.RunIfError(out, err)
	}(c.Globals.ErrLog)

	if err = processSetupCreation(
		newService, so, spinner, c, serviceID, serviceVersion.Number,
	); err != nil {
		return err
	}

	cont, err = processPackage(
		c, pkgPath, serviceID, serviceVersion.Number, spinner, out,
	)
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	if err = processService(c, serviceID, serviceVersion.Number, spinner); err != nil {
		return err
	}

	domain, err := getServiceDomain(c.Globals.APIClient, serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	serviceURL := fmt.Sprintf("https://%s", domain)

	if !c.StatusCheckOff && newService {
		var status int
		if status, err = checkingServiceAvailability(serviceURL+c.StatusCheckPath, spinner, c); err != nil {
			if re, ok := err.(fsterr.RemediationError); ok {
				text.Warning(out, re.Remediation)
			}
		}

		// Because the service availability can return an error (which we ignore),
		// then we need to check for the 'no error' scenarios.
		if err == nil {
			switch {
			case validStatusCodeRange(c.StatusCheckCode) && status != c.StatusCheckCode:
				// If the user set a specific status code expectation...
				text.Warning(out, "The service path `%s` responded with a status code (%d) that didn't match what was expected (%d).", c.StatusCheckPath, status, c.StatusCheckCode)
			case !validStatusCodeRange(c.StatusCheckCode) && status >= http.StatusBadRequest:
				// If no status code was specified, and the actual status response was an error...
				text.Info(out, "The service path `%s` responded with a non-successful status code (%d). Please check your application code if this is an unexpected response.", c.StatusCheckPath, status)
			default:
				text.Break(out)
			}
		}
	}

	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))
	text.Description(out, "View this service at", serviceURL)
	text.Success(out, "Deployed package (service %s, version %v)", serviceID, serviceVersion.Number)
	return nil
}

// validStatusCodeRange checks the status is a valid status code.
// e.g. >= 100 and <= 999
func validStatusCodeRange(status int) bool {
	if status >= 100 && status <= 999 {
		return true
	}
	return false
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
	serviceID, pkgPath string,
	err error,
) {
	defaultActivator := func(customerID string) error { return nil }

	token, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return defaultActivator, 0, "", "", fsterr.ErrNoToken
	}

	// IMPORTANT: We don't handle the error when looking up the Service ID.
	// This is because later in the Exec() flow we might create a 'new' service.
	// Refer to manageNoServiceIDFlow()
	serviceID, source, flag, err := cmd.ServiceID(c.ServiceName, c.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err == nil && c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	pkgPath, err = validatePackage(c.Manifest, c.Package, c.Globals.Verbose(), c.Globals.ErrLog, out)
	if err != nil {
		return defaultActivator, source, serviceID, "", err
	}

	endpoint, _ := c.Globals.Endpoint()
	fnActivateTrial = preconfigureActivateTrial(endpoint, token, c.Globals.HTTPClient)

	return fnActivateTrial, source, serviceID, pkgPath, err
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
) (pkgPath string, err error) {
	err = data.File.ReadError()
	if err != nil {
		if packageFlag == "" {
			if errors.Is(err, os.ErrNotExist) {
				err = fsterr.ErrReadingManifest
			}
			return pkgPath, err
		}

		// NOTE: Before returning the manifest read error, we'll attempt to read
		// the manifest from within the given package archive.
		err := readManifestFromPackageArchive(&data, packageFlag, verbose, out)
		if err != nil {
			return pkgPath, err
		}
	}

	projectName, source := data.Name()
	pkgPath, err = packagePath(packageFlag, projectName, source)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": packageFlag,
		})
		return pkgPath, err
	}

	pkgSize, err := packageSize(pkgPath)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": pkgPath,
		})
		return pkgPath, fsterr.RemediationError{
			Inner:       fmt.Errorf("error reading package size: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}

	if pkgSize > MaxPackageSize {
		return pkgPath, fsterr.RemediationError{
			Inner:       fmt.Errorf("package size is too large (%d bytes)", pkgSize),
			Remediation: fsterr.PackageSizeRemediation,
		}
	}

	if err := validatePackageContent(pkgPath); err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Package path": pkgPath,
			"Package size": pkgSize,
		})
		return pkgPath, err
	}

	return pkgPath, nil
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
		if errors.Is(err, os.ErrNotExist) {
			err = fsterr.ErrReadingManifest
		}
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
		_, err := undocumented.Call(undocumented.CallOptions{
			APIEndpoint: endpoint,
			HTTPClient:  httpClient,
			Method:      http.MethodPost,
			Path:        fmt.Sprintf(undocumented.EdgeComputeTrial, customerID),
			Token:       token,
		})
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
	spinner text.Spinner,
) (newService bool, updatedServiceID string, serviceVersion *fastly.Version, cont bool, err error) {
	if source == manifest.SourceUndefined {
		newService = true
		serviceID, serviceVersion, err = manageNoServiceIDFlow(
			c.Globals.Flags, in, out,
			c.Globals.APIClient, c.Package, c.Globals.ErrLog,
			&c.Manifest.File, fnActivateTrial, spinner, c.ServiceName,
		)
		if err != nil {
			return newService, "", nil, false, err
		}
		if serviceID == "" {
			return newService, "", nil, false, nil // user declined service creation prompt
		}
	} else {
		// There is a scenario where a user already has a Service ID within the
		// fastly.toml manifest but they want to deploy their project to a different
		// service (e.g. deploy to a staging service).
		//
		// In this scenario we end up here because we have found a Service ID in the
		// manifest but if the --service-name flag is set, then we need to ignore
		// what's set in the manifest and instead identify the ID of the service
		// name the user has provided.
		if c.ServiceName.WasSet {
			serviceID, err = c.ServiceName.Parse(c.Globals.APIClient)
			if err != nil {
				return false, "", nil, false, err
			}
		}

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
	apiClient api.Interface,
	packageFlag string,
	errLog fsterr.LogInterface,
	manifestFile *manifest.File,
	fnActivateTrial activator,
	spinner text.Spinner,
	serviceNameFlag cmd.OptionalServiceNameID,
) (serviceID string, serviceVersion *fastly.Version, err error) {
	if !f.AutoYes && !f.NonInteractive {
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the fastly.toml file, otherwise follow the prompts to create a service now.")
		text.Break(out)
		text.Output(out, "Press ^C at any time to quit.")

		if manifestFile.Setup.Defined() {
			text.Info(out, "Processing of the fastly.toml [setup] configuration happens only when there is no existing service. Once a service is created, any further changes to the service or its resources must be made manually.")
		} else {
			text.Break(out)
		}

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

	// The service name will be whatever is set in the --service-name flag.
	// If the flag isn't set, and we're able to prompt, we'll ask the user.
	// If the flag isn't set, and we're non-interactive, we'll use the default.
	if serviceNameFlag.WasSet {
		serviceName = serviceNameFlag.Value
	} else if !f.AcceptDefaults && !f.NonInteractive {
		serviceName, err = text.Input(out, text.BoldYellow(fmt.Sprintf("Service name: [%s] ", defaultServiceName)), in)
		if err != nil || serviceName == "" {
			serviceName = defaultServiceName
		}
	} else {
		serviceName = defaultServiceName
	}

	// There is no service and so we'll do a one time creation of the service
	//
	// NOTE: we're shadowing the `serviceVersion` and `serviceID` variables.
	serviceID, serviceVersion, err = createService(f, serviceName, apiClient, fnActivateTrial, spinner, errLog, out)
	if err != nil {
		errLog.AddWithContext(err, map[string]any{
			"Service name": serviceName,
		})
		return serviceID, serviceVersion, err
	}

	err = updateManifestServiceID(manifestFile, manifest.Filename, serviceID)

	// NOTE: Skip error if --package flag is set.
	//
	// This is because the use of the --package flag suggests the user is not
	// within a project directory. If that is the case, then we don't want the
	// error to be returned because of course there is no manifest to update.
	//
	// If the user does happen to be in a project directory and they use the
	// --package flag, then the above function call to update the manifest will
	// have succeeded and so there will be no error.
	if err != nil && packageFlag == "" {
		errLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return serviceID, serviceVersion, err
	}

	return serviceID, serviceVersion, nil
}

// createService creates a service to associate with the compute package.
//
// NOTE: If the creation of the service fails because the user has not
// activated a free trial, then we'll trigger the trial for their account.
func createService(
	f global.Flags,
	serviceName string,
	apiClient api.Interface,
	fnActivateTrial activator,
	spinner text.Spinner,
	errLog fsterr.LogInterface,
	out io.Writer,
) (serviceID string, serviceVersion *fastly.Version, err error) {
	if !f.AcceptDefaults && !f.NonInteractive {
		text.Break(out)
	}

	err = spinner.Start()
	if err != nil {
		return "", nil, err
	}
	msg := "Creating service"
	spinner.Message(msg + "...")

	service, err := apiClient.CreateService(&fastly.CreateServiceInput{
		Name: &serviceName,
		Type: fastly.String("wasm"),
	})
	if err != nil {
		if strings.Contains(err.Error(), trialNotActivated) {
			user, err := apiClient.GetCurrentUser()
			if err != nil {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return "", nil, spinErr
				}

				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       fmt.Errorf("unable to identify user associated with the given token: %w", err),
					Remediation: "To ensure you have access to the Compute@Edge platform we need your Customer ID. " + fsterr.AuthRemediation,
				}
			}

			err = fnActivateTrial(user.CustomerID)
			if err != nil {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return "", nil, spinErr
				}

				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       fmt.Errorf("error creating service: you do not have the Compute@Edge free trial enabled on your Fastly account"),
					Remediation: fsterr.ComputeTrialRemediation,
				}
			}

			errLog.AddWithContext(err, map[string]any{
				"Service Name": serviceName,
				"Customer ID":  user.CustomerID,
			})

			spinner.StopFailMessage(msg)
			err = spinner.StopFail()
			if err != nil {
				return "", nil, err
			}

			return createService(f, serviceName, apiClient, fnActivateTrial, spinner, errLog, out)
		}

		errLog.AddWithContext(err, map[string]any{
			"Service Name": serviceName,
		})
		return serviceID, serviceVersion, fmt.Errorf("error creating service: %w", err)
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return "", nil, err
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
		return fmt.Errorf("error reading fastly.toml: %w", err)
	}

	m.ServiceID = serviceID

	if err := m.Write(manifestFilename); err != nil {
		return fmt.Errorf("error saving fastly.toml: %w", err)
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

// pkgCompare compares the local package files hash against the existing service
// package version and exits early with message if identical.
func pkgCompare(client api.Interface, serviceID string, version int, filesHash string, out io.Writer) (bool, error) {
	p, err := client.GetPackage(&fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})

	if err == nil {
		if filesHash == p.Metadata.FilesHash {
			text.Info(out, "Skipping package deployment, local and service version are identical. (service %v, version %v) ", serviceID, version)
			return false, nil
		}
	}

	return true, nil
}

// pkgUpload uploads the package to the specified service and version.
func pkgUpload(spinner text.Spinner, client api.Interface, serviceID string, version int, path string) error {
	return spinner.Process("Uploading package", func(_ *text.SpinnerWrapper) error {
		_, err := client.UpdatePackage(&fastly.UpdatePackageInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			PackagePath:    path,
		})
		if err != nil {
			return fmt.Errorf("error uploading package: %w", err)
		}
		return nil
	})
}

// setupObjects is a collection of backend objects created during setup.
// Objects may be nil.
type setupObjects struct {
	domains      *setup.Domains
	backends     *setup.Backends
	configStores *setup.ConfigStores
	loggers      *setup.Loggers
	objectStores *setup.KVStores
	kvStores     *setup.KVStores
	secretStores *setup.SecretStores
}

func constructSetupObjects(
	newService bool,
	serviceID string,
	serviceVersion int,
	c *DeployCommand,
	in io.Reader,
	out io.Writer,
) (setupObjects, error) {
	var (
		so  setupObjects
		err error
	)

	// We only check the Service ID is valid when handling an existing service.
	if !newService {
		if err = checkServiceID(serviceID, c.Globals.APIClient); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return setupObjects{}, err
		}
	}

	// Because a service_id exists in the fastly.toml doesn't mean it's valid
	// e.g. it could be missing required resources such as a domain or backend.
	// We check and allow the user to configure these settings before continuing.

	so.domains = &setup.Domains{
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

	if err = so.domains.Validate(); err != nil {
		errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
		return setupObjects{}, fmt.Errorf("error configuring service domains: %w", err)
	}

	if newService {
		so.backends = &setup.Backends{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.Backends,
			Stdin:          in,
			Stdout:         out,
		}

		so.configStores = &setup.ConfigStores{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.ConfigStores,
			Stdin:          in,
			Stdout:         out,
		}

		so.loggers = &setup.Loggers{
			Setup:  c.Manifest.File.Setup.Loggers,
			Stdout: out,
		}

		so.objectStores = &setup.KVStores{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.ObjectStores,
			Stdin:          in,
			Stdout:         out,
		}

		so.kvStores = &setup.KVStores{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.KVStores,
			Stdin:          in,
			Stdout:         out,
		}

		so.secretStores = &setup.SecretStores{
			APIClient:      c.Globals.APIClient,
			AcceptDefaults: c.Globals.Flags.AcceptDefaults,
			NonInteractive: c.Globals.Flags.NonInteractive,
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
			Setup:          c.Manifest.File.Setup.SecretStores,
			Stdin:          in,
			Stdout:         out,
		}
	}

	return so, nil
}

func processSetupConfig(
	newService bool,
	so setupObjects,
	serviceID string,
	serviceVersion int,
	c *DeployCommand,
) error {
	if so.domains.Missing() {
		if err := so.domains.Configure(); err != nil {
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
		if err := so.backends.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service backends: %w", err)
		}

		if so.configStores.Predefined() {
			if err := so.configStores.Configure(); err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service config stores: %w", err)
			}
		}

		if so.loggers.Predefined() {
			// NOTE: We don't handle errors from the Configure() method because we
			// don't actually do anything other than display a message to the user
			// informing them that they need to create a log endpoint and which
			// provider type they should be. The reason we don't implement logic for
			// creating logging objects is because the API input fields vary
			// significantly between providers.
			_ = so.loggers.Configure()
		}

		if so.objectStores.Predefined() {
			if err := so.objectStores.Configure(); err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service object stores: %w", err)
			}
		}

		if so.kvStores.Predefined() {
			if err := so.kvStores.Configure(); err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service kv stores: %w", err)
			}
		}

		if so.secretStores.Predefined() {
			if err := so.secretStores.Configure(); err != nil {
				errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
				return fmt.Errorf("error configuring service secret stores: %w", err)
			}
		}
	}

	return nil
}

func processSetupCreation(
	newService bool,
	so setupObjects,
	spinner text.Spinner,
	c *DeployCommand,
	serviceID string,
	serviceVersion int,
) error {
	if so.domains.Missing() {
		so.domains.Spinner = spinner

		if err := so.domains.Create(); err != nil {
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
		so.backends.Spinner = spinner
		so.configStores.Spinner = spinner
		so.objectStores.Spinner = spinner
		so.kvStores.Spinner = spinner
		so.secretStores.Spinner = spinner

		if err := so.backends.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := so.configStores.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := so.objectStores.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := so.kvStores.Create(); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Accept defaults": c.Globals.Flags.AcceptDefaults,
				"Auto-yes":        c.Globals.Flags.AutoYes,
				"Non-interactive": c.Globals.Flags.NonInteractive,
				"Service ID":      serviceID,
				"Service Version": serviceVersion,
			})
			return err
		}

		if err := so.secretStores.Create(); err != nil {
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
	pkgPath, serviceID string,
	serviceVersion int,
	spinner text.Spinner,
	out io.Writer,
) (cont bool, err error) {
	filesHash, err := getFilesHash(pkgPath)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path": pkgPath,
		})
		return false, err
	}

	cont, err = pkgCompare(c.Globals.APIClient, serviceID, serviceVersion, filesHash, out)
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

	err = pkgUpload(spinner, c.Globals.APIClient, serviceID, serviceVersion, pkgPath)
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

func processService(c *DeployCommand, serviceID string, serviceVersion int, spinner text.Spinner) error {
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

	return spinner.Process(fmt.Sprintf("Activating service (version %d)", serviceVersion), func(_ *text.SpinnerWrapper) error {
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
	})
}

func getServiceDomain(apiClient api.Interface, serviceID string, serviceVersion int) (string, error) {
	latestDomains, err := apiClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	})
	if err != nil {
		return "", err
	}
	name := latestDomains[0].Name
	if segs := strings.Split(name, "*."); len(segs) > 1 {
		name = segs[1]
	}
	return name, nil
}

// checkingServiceAvailability pings the service URL until either there is a
// non-500 (or whatever status code is configured by the user) or if the
// configured timeout is reached.
func checkingServiceAvailability(
	serviceURL string,
	spinner text.Spinner,
	c *DeployCommand,
) (status int, err error) {
	remediation := "The service has been successfully deployed and activated, but the service 'availability' check %s (we were looking for a %s but the last status code response was: %d). If using a custom domain, please be sure to check your DNS settings. Otherwise, your application might be taking longer than usual to deploy across our global network. Please continue to check the service URL and if still unavailable please contact Fastly support."

	dur := time.Duration(c.StatusCheckTimeout) * time.Second
	end := time.Now().Add(dur)
	timeout := time.After(dur)
	ticker := time.NewTicker(1 * time.Second)
	defer func() { ticker.Stop() }()

	err = spinner.Start()
	if err != nil {
		return 0, err
	}
	msg := "Checking service availability"
	spinner.Message(msg + generateTimeout(time.Until(end)))

	expected := "non-500 status code"
	if validStatusCodeRange(c.StatusCheckCode) {
		expected = fmt.Sprintf("%d status code", c.StatusCheckCode)
	}

	// Keep trying until we're timed out, got a result or got an error
	for {
		select {
		case <-timeout:
			returnedStatus := fmt.Sprintf(" (status: %d)", status)
			spinner.StopFailMessage(msg + returnedStatus)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return status, spinErr
			}
			return status, fsterr.RemediationError{
				Inner:       errors.New("service not yet available"),
				Remediation: fmt.Sprintf(remediation, "timed out", expected, status),
			}
		case t := <-ticker.C:
			var (
				ok  bool
				err error
			)
			// We overwrite the `status` variable in the parent scope (defined in the
			// return arguments list) so it can be used as part of both the timeout
			// and success scenarios.
			ok, status, err = pingServiceURL(serviceURL, c.Globals.HTTPClient, c.StatusCheckCode)
			if err != nil {
				returnedStatus := fmt.Sprintf(" (status: %d)", status)
				spinner.StopFailMessage(msg + returnedStatus)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return status, spinErr
				}
				return status, fsterr.RemediationError{
					Inner:       err,
					Remediation: fmt.Sprintf(remediation, "failed", expected, status),
				}
			} else if ok {
				returnedStatus := fmt.Sprintf(" (status: %d)", status)
				spinner.StopMessage(msg + returnedStatus)
				return status, spinner.Stop()
			}
			// Service not available, and no error, so jump back to top of loop
			spinner.Message(msg + generateTimeout(end.Sub(t)))
		}
	}
}

// generateTimeout inserts a dynamically generated message on each tick.
// It notifies the user what's happening and how long is left on the timer.
func generateTimeout(d time.Duration) string {
	remaining := fmt.Sprintf("timeout: %v", d.Round(time.Second))
	return fmt.Sprintf(" (app deploying across Fastly's global network | %s)...", remaining)
}

// pingServiceURL indicates if the service returned a non-5xx response (or
// whatever the user defined with --status-check-code), which should help
// signify if the service is generally available.
func pingServiceURL(serviceURL string, httpClient api.HTTPClient, expectedStatusCode int) (ok bool, status int, err error) {
	req, err := http.NewRequest("GET", serviceURL, nil)
	if err != nil {
		return false, 0, err
	}

	// gosec flagged this:
	// G107 (CWE-88): Potential HTTP request made with variable url
	// Disabling as we trust the source of the variable.
	// #nosec
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// We check for the user's defined status code expectation.
	// Otherwise we'll default to checking for a non-500.
	if validStatusCodeRange(expectedStatusCode) && resp.StatusCode == expectedStatusCode {
		return true, resp.StatusCode, nil
	} else if resp.StatusCode < http.StatusInternalServerError {
		return true, resp.StatusCode, nil
	}
	return false, resp.StatusCode, nil
}
