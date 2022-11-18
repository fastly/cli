package compute

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/kennygrant/sanitize"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	path           string
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update a package on a Fastly Compute@Edge service version")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.path)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return fsterr.ErrNoToken
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
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
		projectName, source := c.manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to read project name: %w", fsterr.ErrReadingManifest),
				Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
			}
		}
		packagePath = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
	}

	progress := text.NewProgress(out, c.Globals.Verbose())
	defer func() {
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			progress.Fail() // progress.Done is handled inline
		}
	}()

	progress.Step("Uploading package...")
	_, err = c.Globals.APIClient.UpdatePackage(&fastly.UpdatePackageInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
		PackagePath:    packagePath,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("error uploading package: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}
	progress.Done()

	text.Success(out, "Updated package (service %s, version %v)", serviceID, serviceVersion.Number)
	return nil
}
