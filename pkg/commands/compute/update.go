package compute

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kennygrant/sanitize"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	argparser.Base
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	Version        string                          `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
	serviceVersion argparser.OptionalServiceVersion
	AutoClone      argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Path           string                      `name:"package" short:"p" help:"Path to a package tar.gz"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	packagePath := c.Path
	if packagePath == "" {
		if err := c.Globals.Manifest.File.ReadError(); err != nil {
			err = packageFlagManifestError(err)
			c.Globals.ErrLog.Add(err)
			return err
		}
		projectName, source := c.Globals.Manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.ErrMissingManifestName
		}
		packagePath = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
		}
	}()

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	err = spinner.Process("Uploading package", func(_ *text.SpinnerWrapper) error {
		_, err = c.Globals.APIClient.UpdatePackage(context.TODO(), &fastly.UpdatePackageInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersionNumber,
			PackagePath:    new(packagePath),
		})
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("error uploading package: %w", err),
				Remediation: "Run `fastly compute build` to produce a Compute package, alternatively use the --package flag to reference a package outside of the current project.",
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	text.Success(out, "\nUpdated package (service %s, version %v)", serviceID, serviceVersionNumber)
	return nil
}
