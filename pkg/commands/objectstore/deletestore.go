package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// DeleteStoreCommand calls the Fastly API to delete an object store.
type DeleteStoreCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteObjectStoreInput
}

// NewDeleteStoreCommand returns a usable command registered under the parent.
func NewDeleteStoreCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteStoreCommand {
	c := DeleteStoreCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("delete", "Delete an object store")
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteStoreCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.DeleteObjectStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Deleted object store ID %s", c.Input.ID)
	return nil
}
