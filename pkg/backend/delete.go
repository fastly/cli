package backend

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// DeleteCommand calls the Fastly API to delete backends.
type DeleteCommand struct {
	common.Base
	Input fastly.DeleteBackendInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent common.Registerer, globals *config.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.CmdClause = parent.Command("delete", "Delete a backend on a Fastly service version").Alias("remove")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.Service)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "Backend name").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if err := c.Globals.Client.DeleteBackend(&c.Input); err != nil {
		return err
	}

	text.Success(out, "Deleted backend %s (service %s version %d)", c.Input.Name, c.Input.Service, c.Input.Version)
	return nil
}
