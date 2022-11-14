package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// DeleteCommand calls the Fastly API to delete an object store.
type DeleteCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteObjectStoreInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("delete", "Delete a Fastly object store")
	c.CmdClause.Flag("id", "ID of object store").Required().StringVar(&c.Input.ID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.DeleteObjectStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Deleted object store ID:%s", c.Input.ID)
	return nil
}
