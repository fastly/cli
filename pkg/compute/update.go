package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	path           string
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a package on a Fastly Compute@Edge service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.SetAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("path", "Path to package").Required().Short('p').StringVar(&c.path)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(c.manifest, c.serviceVersion, c.autoClone, c.Globals.Flag.Verbose, out, c.Globals.Client)
	if err != nil {
		return err
	}

	progress := text.NewQuietProgress(out)
	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	progress.Step("Uploading package...")
	_, err = c.Globals.Client.UpdatePackage(&fastly.UpdatePackageInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
		PackagePath:    c.path,
	})
	if err != nil {
		return fmt.Errorf("error uploading package: %w", err)
	}
	progress.Done()

	text.Success(out, "Updated package (service %s, version %v)", serviceID, serviceVersion.Number)
	return nil
}
