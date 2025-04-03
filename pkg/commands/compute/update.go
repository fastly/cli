package compute

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/kennygrant/sanitize"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	argparser.Base
	path           string
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	autoClone      argparser.OptionalAutoClone
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a package on a Fastly Compute service version")
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.path)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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

	packagePath := c.path
	if packagePath == "" {
		projectName, source := c.Globals.Manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to read project name: %w", fsterr.ErrReadingManifest),
				Remediation: "Run `fastly compute build` to produce a Compute package, alternatively use the --package flag to reference a package outside of the current project.",
			}
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
		_, err = c.Globals.APIClient.UpdatePackage(&fastly.UpdatePackageInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersionNumber,
			PackagePath:    fastly.ToPointer(packagePath),
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
