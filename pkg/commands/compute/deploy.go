package compute

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

// ErrPackageUnchanged is an error that indicates the package hasn't changed.
var ErrPackageUnchanged = errors.New("package is unchanged")

// DeployCommand deploys an artifact previously produced by build.
type DeployCommand struct {
	cmd.Base
	manifestPath string

	// NOTE: these are public so that the "publish" composite command can set the
	// values appropriately before calling the Exec() function.
	Comment            cmd.OptionalString
	Dir                string
	Domain             string
	Env                string
	PackagePath        string
	ServiceName        cmd.OptionalServiceNameID
	ServiceVersion     cmd.OptionalServiceVersion
	StatusCheckCode    int
	StatusCheckOff     bool
	StatusCheckPath    string
	StatusCheckTimeout int
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent cmd.Registerer, g *global.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = g
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute service")

	// NOTE: when updating these flags, be sure to update the composite command:
	// `compute publish`.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Globals.Manifest.Flag.ServiceID,
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
	c.CmdClause.Flag("dir", "Project directory (default: current directory)").Short('C').StringVar(&c.Dir)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.Domain)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").StringVar(&c.Env)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.PackagePath)
	c.CmdClause.Flag("status-check-code", "Set the expected status response for the service availability check").IntVar(&c.StatusCheckCode)
	c.CmdClause.Flag("status-check-off", "Disable the service availability check").BoolVar(&c.StatusCheckOff)
	c.CmdClause.Flag("status-check-path", "Specify the URL path for the service availability check").Default("/").StringVar(&c.StatusCheckPath)
	c.CmdClause.Flag("status-check-timeout", "Set a timeout (in seconds) for the service availability check").Default("120").IntVar(&c.StatusCheckTimeout)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	manifestFilename := EnvironmentManifest(c.Env)
	if c.Env != "" {
		if c.Globals.Verbose() {
			text.Info(out, EnvManifestMsg, manifestFilename, manifest.Filename)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	c.manifestPath = filepath.Join(wd, manifestFilename)

	projectDir, err := ChangeProjectDirectory(c.Dir)
	if err != nil {
		return err
	}
	if projectDir != "" {
		if c.Globals.Verbose() {
			text.Info(out, ProjectDirMsg, projectDir)
		}
		c.manifestPath = filepath.Join(projectDir, manifestFilename)
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	err = spinner.Process(fmt.Sprintf("Verifying %s", manifestFilename), func(_ *text.SpinnerWrapper) error {
		if projectDir != "" || c.Env != "" {
			err = c.Globals.Manifest.File.Read(c.manifestPath)
		} else {
			err = c.Globals.Manifest.File.ReadError()
		}
		if err != nil {
			// If the user hasn't specified a package to deploy, then we'll just check
			// the read error and return it.
			if c.PackagePath == "" {
				if errors.Is(err, os.ErrNotExist) {
					err = fsterr.ErrReadingManifest
				}
				c.Globals.ErrLog.Add(err)
				return err
			}
			// Otherwise, we'll attempt to read the manifest from within the given
			// package archive.
			if err := readManifestFromPackageArchive(&c.Globals.Manifest, c.PackagePath, manifestFilename); err != nil {
				return err
			}
			if c.Globals.Verbose() {
				text.Info(out, "Using %s within --package archive: %s\n\n", manifestFilename, c.PackagePath)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !c.Globals.Flags.NonInteractive {
		text.Break(out)
	}

	fnActivateTrial, serviceID, err := c.Setup(out)
	if err != nil {
		return err
	}
	noExistingService := serviceID == ""

	undoStack := undo.NewStack()
	undoStack.Push(func() error {
		if noExistingService && serviceID != "" {
			return c.CleanupNewService(serviceID, manifestFilename, out)
		}
		return nil
	})

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
		}
		undoStack.RunIfError(out, err)
	}(c.Globals.ErrLog)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go monitorSignals(signalCh, noExistingService, out, undoStack, spinner)

	var serviceVersion *fastly.Version
	if noExistingService {
		serviceID, serviceVersion, err = c.NewService(manifestFilename, fnActivateTrial, spinner, in, out)
		if err != nil {
			return err
		}
		if serviceID == "" {
			return nil // user declined service creation prompt
		}
	} else {
		serviceVersion, err = c.ExistingServiceVersion(serviceID, out)
		if err != nil {
			if errors.Is(err, ErrPackageUnchanged) {
				text.Info(out, "Skipping package deployment, local and service version are identical. (service %s, version %d) ", serviceID, serviceVersion.Number)
				return nil
			}
			return err
		}
		if c.Globals.Manifest.File.Setup.Defined() && !c.Globals.Flags.Quiet {
			text.Info(out, "\nProcessing of the %s [setup] configuration happens only for a new service. Once a service is created, any further changes to the service or its resources must be made manually.\n\n", manifestFilename)
		}
	}

	var sr ServiceResources

	// NOTE: A 'domain' resource isn't strictly part of the [setup] config.
	// It's part of the implementation so that we can utilise the same interface.
	// A domain is required regardless of whether it's a new service or existing.
	sr.domains = &setup.Domains{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		PackageDomain:  c.Domain,
		RetryLimit:     5,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
		Stdin:          in,
		Stdout:         out,
		Verbose:        c.Globals.Verbose(),
	}
	if err = sr.domains.Validate(); err != nil {
		errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion.Number)
		return fmt.Errorf("error configuring service domains: %w", err)
	}
	if noExistingService {
		c.ConstructNewServiceResources(
			&sr, serviceID, serviceVersion.Number, in, out,
		)
	}

	if sr.domains.Missing() {
		if err := sr.domains.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion.Number)
			return fmt.Errorf("error configuring service domains: %w", err)
		}
	}
	if noExistingService {
		if err = c.ConfigureServiceResources(sr, serviceID, serviceVersion.Number); err != nil {
			return err
		}
	}

	if sr.domains.Missing() {
		sr.domains.Spinner = spinner
		if err := sr.domains.Create(); err != nil {
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
	if noExistingService {
		if err = c.CreateServiceResources(sr, spinner, serviceID, serviceVersion.Number); err != nil {
			return err
		}
	}

	err = c.UploadPackage(spinner, serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path":    c.PackagePath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if err = c.ProcessService(serviceID, serviceVersion.Number, spinner); err != nil {
		return err
	}

	serviceURL, err := c.GetServiceURL(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	if !c.StatusCheckOff && noExistingService {
		c.StatusCheck(serviceURL, spinner, out)
	}

	if !noExistingService {
		text.Break(out)
	}
	displayDeployOutput(out, manageServiceBaseURL, serviceID, serviceURL, serviceVersion.Number)
	return nil
}

// StatusCheck checks the service URL and identifies when it's ready.
func (c *DeployCommand) StatusCheck(serviceURL string, spinner text.Spinner, out io.Writer) {
	var (
		err    error
		status int
	)
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

func displayDeployOutput(out io.Writer, manageServiceBaseURL, serviceID, serviceURL string, serviceVersion int) {
	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))
	text.Description(out, "View this service at", serviceURL)
	text.Success(out, "Deployed package (service %s, version %v)", serviceID, serviceVersion)
}

// validStatusCodeRange checks the status is a valid status code.
// e.g. >= 100 and <= 999.
func validStatusCodeRange(status int) bool {
	if status >= 100 && status <= 999 {
		return true
	}
	return false
}

// Setup prepares the environment.
//
// - Check if there is an API token missing.
// - Acquire the Service ID/Version.
// - Validate there is a package to deploy.
// - Determine if a trial needs to be activated on the user's account.
func (c *DeployCommand) Setup(out io.Writer) (fnActivateTrial Activator, serviceID string, err error) {
	defaultActivator := func(customerID string) error { return nil }

	token, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return defaultActivator, "", fsterr.ErrNoToken
	}

	// IMPORTANT: We don't handle the error when looking up the Service ID.
	// This is because later in the Exec() flow we might create a 'new' service.
	serviceID, source, flag, err := cmd.ServiceID(c.ServiceName, c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err == nil && c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	if c.PackagePath == "" {
		projectName, source := c.Globals.Manifest.Name()
		if source == manifest.SourceUndefined {
			return defaultActivator, serviceID, fsterr.ErrReadingManifest
		}
		c.PackagePath = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
	}

	err = validatePackage(c.PackagePath)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path": c.PackagePath,
		})
		return defaultActivator, serviceID, err
	}

	endpoint, _ := c.Globals.Endpoint()
	fnActivateTrial = preconfigureActivateTrial(endpoint, token, c.Globals.HTTPClient)

	return fnActivateTrial, serviceID, err
}

// validatePackage checks the package and returns its path, which can change
// depending on the user flow scenario.
func validatePackage(pkgPath string) error {
	pkgSize, err := packageSize(pkgPath)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("error reading package size: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}
	if pkgSize > MaxPackageSize {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("package size is too large (%d bytes)", pkgSize),
			Remediation: fsterr.PackageSizeRemediation,
		}
	}
	return validatePackageContent(pkgPath)
}

// readManifestFromPackageArchive extracts the manifest file from the given
// package archive file and reads it into memory.
func readManifestFromPackageArchive(data *manifest.Data, packageFlag, manifestFilename string) error {
	dst, err := os.MkdirTemp("", fmt.Sprintf("%s-*", manifestFilename))
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

	manifestPath, err := locateManifest(filepath.Join(dst, extractedDirName), manifestFilename)
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

	return nil
}

// locateManifest attempts to find the manifest within the given path's
// directory tree.
func locateManifest(path, manifestFilename string) (string, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var foundManifest string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Base(path) == manifestFilename {
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

// packageSize returns the size of the .tar.gz package.
//
// Reference:
// https://docs.fastly.com/products/compute-at-edge-billing-and-resource-limits#resource-limits
func packageSize(path string) (size int64, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return size, err
	}
	return fi.Size(), nil
}

// Activator represents a function that calls an undocumented API endpoint for
// activating a Compute free trial on the given customer account.
//
// It is preconfigured with the Fastly API endpoint, a user token and a simple
// HTTP Client.
//
// This design allows us to pass an Activator rather than passing multiple
// unrelated arguments through several nested functions.
type Activator func(customerID string) error

// preconfigureActivateTrial activates a free trial on the customer account.
func preconfigureActivateTrial(endpoint, token string, httpClient api.HTTPClient) Activator {
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
			// 409 Conflict == The Compute trial has already been created.
			if apiErr.StatusCode != http.StatusConflict {
				return fmt.Errorf("%w: %d %s", err, apiErr.StatusCode, http.StatusText(apiErr.StatusCode))
			}
		}
		return nil
	}
}

// NewService handles creating a new service when no Service ID is found.
func (c *DeployCommand) NewService(manifestFilename string, fnActivateTrial Activator, spinner text.Spinner, in io.Reader, out io.Writer) (string, *fastly.Version, error) {
	var (
		err            error
		serviceID      string
		serviceVersion *fastly.Version
	)

	if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Output(out, "There is no Fastly service associated with this package. To connect to an existing service add the Service ID to the %s file, otherwise follow the prompts to create a service now.\n\n", manifestFilename)
		text.Output(out, "Press ^C at any time to quit.")

		if c.Globals.Manifest.File.Setup.Defined() {
			text.Info(out, "\nProcessing of the %s [setup] configuration happens only when there is no existing service. Once a service is created, any further changes to the service or its resources must be made manually.", manifestFilename)
		}

		text.Break(out)
		answer, err := text.AskYesNo(out, "Create new service: [y/N] ", in)
		if err != nil {
			return serviceID, serviceVersion, err
		}
		if !answer {
			return serviceID, serviceVersion, nil
		}
		text.Break(out)
	}

	defaultServiceName := c.Globals.Manifest.File.Name
	var serviceName string

	// The service name will be whatever is set in the --service-name flag.
	// If the flag isn't set, and we're non-interactive, we'll use the default.
	// If the flag isn't set, and we're interactive, we'll prompt the user.
	switch {
	case c.ServiceName.WasSet:
		serviceName = c.ServiceName.Value
	case c.Globals.Flags.AcceptDefaults || c.Globals.Flags.NonInteractive:
		serviceName = defaultServiceName
	default:
		serviceName, err = text.Input(out, text.Prompt(fmt.Sprintf("Service name: [%s] ", defaultServiceName)), in)
		if err != nil || serviceName == "" {
			serviceName = defaultServiceName
		}
	}

	// There is no service and so we'll do a one time creation of the service
	//
	// NOTE: we're shadowing the `serviceID` and `serviceVersion` variables.
	serviceID, serviceVersion, err = createService(c.Globals, serviceName, fnActivateTrial, spinner, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service name": serviceName,
		})
		return serviceID, serviceVersion, err
	}

	err = c.UpdateManifestServiceID(serviceID, c.manifestPath)

	// NOTE: Skip error if --package flag is set.
	//
	// This is because the use of the --package flag suggests the user is not
	// within a project directory. If that is the case, then we don't want the
	// error to be returned because of course there is no manifest to update.
	//
	// If the user does happen to be in a project directory and they use the
	// --package flag, then the above function call to update the manifest will
	// have succeeded and so there will be no error.
	if err != nil && c.PackagePath == "" {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
	g *global.Data,
	serviceName string,
	fnActivateTrial Activator,
	spinner text.Spinner,
	out io.Writer,
) (serviceID string, serviceVersion *fastly.Version, err error) {
	f := g.Flags
	apiClient := g.APIClient
	errLog := g.ErrLog

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
				err = fmt.Errorf("unable to identify user associated with the given token: %w", err)
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return "", nil, fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       err,
					Remediation: "To ensure you have access to the Compute platform we need your Customer ID. " + fsterr.AuthRemediation,
				}
			}

			err = fnActivateTrial(user.CustomerID)
			if err != nil {
				err = fmt.Errorf("error creating service: you do not have the Compute free trial enabled on your Fastly account")
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return "", nil, fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				return serviceID, serviceVersion, fsterr.RemediationError{
					Inner:       err,
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

			return createService(g, serviceName, fnActivateTrial, spinner, out)
		}

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return "", nil, spinErr
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

// CleanupNewService is executed if a new service flow has errors.
// It deletes the service, which will cause any contained resources to be deleted.
// It will also strip the Service ID from the fastly.toml manifest file.
func (c *DeployCommand) CleanupNewService(serviceID, manifestFilename string, out io.Writer) error {
	text.Info(out, "\nCleaning up service\n\n")
	err := c.Globals.APIClient.DeleteService(&fastly.DeleteServiceInput{
		ID: serviceID,
	})
	if err != nil {
		return err
	}

	text.Info(out, "Removing Service ID from %s\n\n", manifestFilename)
	err = c.UpdateManifestServiceID("", c.manifestPath)
	if err != nil {
		return err
	}

	text.Output(out, "Cleanup complete")
	return nil
}

// UpdateManifestServiceID updates the Service ID in the manifest.
//
// There are two scenarios where this function is called. The first is when we
// have a Service ID to insert into the manifest. The other is when there is an
// error in the deploy flow, and for which the Service ID will be set to an
// empty string (otherwise the service itself will be deleted while the
// manifest will continue to hold a reference to it).
func (c *DeployCommand) UpdateManifestServiceID(serviceID, manifestPath string) error {
	if err := c.Globals.Manifest.File.Read(manifestPath); err != nil {
		return fmt.Errorf("error reading %s: %w", manifestPath, err)
	}
	c.Globals.Manifest.File.ServiceID = serviceID
	if err := c.Globals.Manifest.File.Write(manifestPath); err != nil {
		return fmt.Errorf("error saving %s: %w", manifestPath, err)
	}
	return nil
}

// errLogService records the error, service id and version into the error log.
func errLogService(l fsterr.LogInterface, err error, sid string, sv int) {
	l.AddWithContext(err, map[string]any{
		"Service ID":      sid,
		"Service Version": sv,
	})
}

// CompareLocalRemotePackage compares the local package files hash against the
// existing service package version and exits early with message if identical.
//
// NOTE: We can't avoid the first 'no-changes' upload after the initial deploy.
// This is because the fastly.toml manifest does actual change after first deploy.
// When user first deploys, there is no value for service_id.
// That version of the manifest is inside the package we're checking against.
// So on the second deploy, even if user has made no changes themselves, we will
// still upload that package because technically there was a change made by the
// CLI to add the Service ID. Any subsequent deploys will be aborted because
// there will be no changes made by the CLI nor the user.
func (c *DeployCommand) CompareLocalRemotePackage(serviceID string, version int) error {
	filesHash, err := getFilesHash(c.PackagePath)
	if err != nil {
		return err
	}
	p, err := c.Globals.APIClient.GetPackage(&fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	// IMPORTANT: Skip error as some services won't have a package to compare.
	// This happens in situations where a user will create the service outside of
	// the CLI and then reference the Service ID in their fastly.toml manifest.
	// In that scenario the service might just be an empty service and so trying
	// to get the package from the service with 404.
	if err == nil && filesHash == p.Metadata.FilesHash {
		return ErrPackageUnchanged
	}
	return nil
}

// UploadPackage uploads the package to the specified service and version.
func (c *DeployCommand) UploadPackage(spinner text.Spinner, serviceID string, version int) error {
	return spinner.Process("Uploading package", func(_ *text.SpinnerWrapper) error {
		_, err := c.Globals.APIClient.UpdatePackage(&fastly.UpdatePackageInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			PackagePath:    c.PackagePath,
		})
		if err != nil {
			return fmt.Errorf("error uploading package: %w", err)
		}
		return nil
	})
}

// ServiceResources is a collection of backend objects created during setup.
// Objects may be nil.
type ServiceResources struct {
	domains      *setup.Domains
	backends     *setup.Backends
	configStores *setup.ConfigStores
	loggers      *setup.Loggers
	objectStores *setup.KVStores
	kvStores     *setup.KVStores
	secretStores *setup.SecretStores
}

// ConstructNewServiceResources instantiates multiple [setup] config resources for a
// new Service to process.
func (c *DeployCommand) ConstructNewServiceResources(
	sr *ServiceResources,
	serviceID string,
	serviceVersion int,
	in io.Reader,
	out io.Writer,
) {
	sr.backends = &setup.Backends{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Setup:          c.Globals.Manifest.File.Setup.Backends,
		Stdin:          in,
		Stdout:         out,
	}

	sr.configStores = &setup.ConfigStores{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Setup:          c.Globals.Manifest.File.Setup.ConfigStores,
		Stdin:          in,
		Stdout:         out,
	}

	sr.loggers = &setup.Loggers{
		Setup:  c.Globals.Manifest.File.Setup.Loggers,
		Stdout: out,
	}

	sr.objectStores = &setup.KVStores{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Setup:          c.Globals.Manifest.File.Setup.ObjectStores,
		Stdin:          in,
		Stdout:         out,
	}

	sr.kvStores = &setup.KVStores{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Setup:          c.Globals.Manifest.File.Setup.KVStores,
		Stdin:          in,
		Stdout:         out,
	}

	sr.secretStores = &setup.SecretStores{
		APIClient:      c.Globals.APIClient,
		AcceptDefaults: c.Globals.Flags.AcceptDefaults,
		NonInteractive: c.Globals.Flags.NonInteractive,
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Setup:          c.Globals.Manifest.File.Setup.SecretStores,
		Stdin:          in,
		Stdout:         out,
	}
}

// ConfigureServiceResources calls the .Predefined() and .Configure() methods
// for each [setup] resource, which first checks if a [setup] config has been
// defined for the resource type, and if so it prompts the user for details.
func (c *DeployCommand) ConfigureServiceResources(sr ServiceResources, serviceID string, serviceVersion int) error {
	// NOTE: A service can't be activated without at least one backend defined.
	// This explains why the following block of code isn't wrapped in a call to
	// the .Predefined() method, as the call to .Configure() will ensure the
	// user is prompted regardless of whether there is a [setup.backends]
	// defined in the fastly.toml configuration.
	if err := sr.backends.Configure(); err != nil {
		errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
		return fmt.Errorf("error configuring service backends: %w", err)
	}

	if sr.configStores.Predefined() {
		if err := sr.configStores.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service config stores: %w", err)
		}
	}

	if sr.loggers.Predefined() {
		// NOTE: We don't handle errors from the Configure() method because we
		// don't actually do anything other than display a message to the user
		// informing them that they need to create a log endpoint and which
		// provider type they should be. The reason we don't implement logic for
		// creating logging objects is because the API input fields vary
		// significantly between providers.
		_ = sr.loggers.Configure()
	}

	if sr.objectStores.Predefined() {
		if err := sr.objectStores.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service object stores: %w", err)
		}
	}

	if sr.kvStores.Predefined() {
		if err := sr.kvStores.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service kv stores: %w", err)
		}
	}

	if sr.secretStores.Predefined() {
		if err := sr.secretStores.Configure(); err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion)
			return fmt.Errorf("error configuring service secret stores: %w", err)
		}
	}

	return nil
}

// CreateServiceResources makes API calls to create resources that have been
// defined in the fastly.toml [setup] configuration.
func (c *DeployCommand) CreateServiceResources(
	sr ServiceResources,
	spinner text.Spinner,
	serviceID string,
	serviceVersion int,
) error {
	sr.backends.Spinner = spinner
	sr.configStores.Spinner = spinner
	sr.objectStores.Spinner = spinner
	sr.kvStores.Spinner = spinner
	sr.secretStores.Spinner = spinner

	if err := sr.backends.Create(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Accept defaults": c.Globals.Flags.AcceptDefaults,
			"Auto-yes":        c.Globals.Flags.AutoYes,
			"Non-interactive": c.Globals.Flags.NonInteractive,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if err := sr.configStores.Create(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Accept defaults": c.Globals.Flags.AcceptDefaults,
			"Auto-yes":        c.Globals.Flags.AutoYes,
			"Non-interactive": c.Globals.Flags.NonInteractive,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if err := sr.objectStores.Create(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Accept defaults": c.Globals.Flags.AcceptDefaults,
			"Auto-yes":        c.Globals.Flags.AutoYes,
			"Non-interactive": c.Globals.Flags.NonInteractive,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if err := sr.kvStores.Create(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Accept defaults": c.Globals.Flags.AcceptDefaults,
			"Auto-yes":        c.Globals.Flags.AutoYes,
			"Non-interactive": c.Globals.Flags.NonInteractive,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if err := sr.secretStores.Create(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Accept defaults": c.Globals.Flags.AcceptDefaults,
			"Auto-yes":        c.Globals.Flags.AutoYes,
			"Non-interactive": c.Globals.Flags.NonInteractive,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	return nil
}

// ProcessService updates the service version comment and then activates the
// service version.
func (c *DeployCommand) ProcessService(serviceID string, serviceVersion int, spinner text.Spinner) error {
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

// GetServiceURL returns the service URL.
func (c *DeployCommand) GetServiceURL(serviceID string, serviceVersion int) (string, error) {
	latestDomains, err := c.Globals.APIClient.ListDomains(&fastly.ListDomainsInput{
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
	return fmt.Sprintf("https://%s", name), nil
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
			err := errors.New("timeout: service not yet available")
			returnedStatus := fmt.Sprintf(" (status: %d)", status)
			spinner.StopFailMessage(msg + returnedStatus)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return status, fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return status, fsterr.RemediationError{
				Inner:       err,
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
				err := fmt.Errorf("failed to ping service URL: %w", err)
				returnedStatus := fmt.Sprintf(" (status: %d)", status)
				spinner.StopFailMessage(msg + returnedStatus)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return status, fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				return status, fsterr.RemediationError{
					Inner:       err,
					Remediation: fmt.Sprintf(remediation, "failed", expected, status),
				}
			}
			if ok {
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

// ExistingServiceVersion returns a Service Version for an existing service.
// If the current service version is active or locked, we clone the version.
func (c *DeployCommand) ExistingServiceVersion(serviceID string, out io.Writer) (*fastly.Version, error) {
	var (
		err            error
		serviceVersion *fastly.Version
	)

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
			return nil, err
		}
	}

	serviceVersion, err = c.ServiceVersion.Parse(serviceID, c.Globals.APIClient)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path": c.PackagePath,
			"Service ID":   serviceID,
		})
		return nil, err
	}

	// Validate that we're dealing with a Compute 'wasm' service and not a
	// VCL service, for which we cannot upload a wasm package format to.
	serviceDetails, err := c.Globals.APIClient.GetServiceDetails(&fastly.GetServiceInput{ID: serviceID})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return serviceVersion, err
	}
	if serviceDetails.Type != "wasm" {
		c.Globals.ErrLog.AddWithContext(fmt.Errorf("error: invalid service type: '%s'", serviceDetails.Type), map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
			"Service Type":    serviceDetails.Type,
		})
		return serviceVersion, fsterr.RemediationError{
			Inner:       fmt.Errorf("invalid service type: %s", serviceDetails.Type),
			Remediation: "Ensure the provided Service ID is associated with a 'Wasm' Fastly Service and not a 'VCL' Fastly service. " + fsterr.ComputeTrialRemediation,
		}
	}

	err = c.CompareLocalRemotePackage(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Package path":    c.PackagePath,
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return serviceVersion, err
	}

	// Unlike other CLI commands that are a direct mapping to an API endpoint,
	// the compute deploy command is a composite of behaviours, and so as we
	// already automatically activate a version we should autoclone without
	// requiring the user to explicitly provide an --autoclone flag.
	if serviceVersion.Active || serviceVersion.Locked {
		clonedVersion, err := c.Globals.APIClient.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion.Number,
		})
		if err != nil {
			errLogService(c.Globals.ErrLog, err, serviceID, serviceVersion.Number)
			return serviceVersion, fmt.Errorf("error cloning service version: %w", err)
		}
		if c.Globals.Verbose() {
			msg := "Service version %d is not editable, so it was automatically cloned. Now operating on version %d.\n\n"
			format := fmt.Sprintf(msg, serviceVersion.Number, clonedVersion.Number)
			text.Output(out, format)
		}
		serviceVersion = clonedVersion
	}

	return serviceVersion, nil
}

func monitorSignals(signalCh chan os.Signal, noExistingService bool, out io.Writer, undoStack *undo.Stack, spinner text.Spinner) {
	<-signalCh
	signal.Stop(signalCh)
	spinner.StopFailMessage("Signal received to interrupt/terminate the Fastly CLI process")
	_ = spinner.StopFail()
	text.Important(out, "\n\nThe Fastly CLI process will be terminated after any clean-up tasks have been processed")
	if noExistingService {
		undoStack.Unwind(out)
	}
	os.Exit(1)
}
