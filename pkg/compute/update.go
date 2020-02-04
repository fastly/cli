package compute

import (
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update packages.
type UpdateCommand struct {
	common.Base
	client    api.HTTPClient
	serviceID string
	version   int
	path      string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.client = client
	c.CmdClause = parent.Command("update", "Update a package on a Fastly Compute@Edge service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.serviceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.version)
	c.CmdClause.Flag("path", "Path to package").Required().Short('p').StringVar(&c.path)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	progress := text.NewQuietProgress(out)
	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	token, source := c.Globals.Token()
	if source == config.SourceUndefined {
		return errors.ErrNoToken
	}
	endpoint, _ := c.Globals.Endpoint()

	progress.Step("Uploading package...")
	client := NewClient(c.client, endpoint, token)
	if err := client.UpdatePackage(c.serviceID, c.version, c.path); err != nil {
		return err
	}
	progress.Done()

	text.Success(out, "Updated package (service %s, version %v)", c.serviceID, c.version)
	return nil
}
