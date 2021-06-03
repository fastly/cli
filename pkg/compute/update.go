package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	cmd.Base
	serviceID string
	version   int
	path      string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update a package on a Fastly Compute@Edge service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.serviceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.version)
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

	progress := text.NewQuietProgress(out)
	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	progress.Step("Uploading package...")
	_, err = c.Globals.Client.UpdatePackage(&fastly.UpdatePackageInput{
		ServiceID:      c.serviceID,
		ServiceVersion: c.version,
		PackagePath:    c.path,
	})
	if err != nil {
		return fmt.Errorf("error uploading package: %w", err)
	}
	progress.Done()

	text.Success(out, "Updated package (service %s, version %v)", c.serviceID, c.version)
	return nil
}
