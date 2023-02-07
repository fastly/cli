package objectstoreentry

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
	Input    fastly.DeleteObjectStoreKeyInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("delete", "Delete a key")
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key-name", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.DeleteObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Deleted key %s from store ID %s", c.Input.Key, c.Input.ID)
	return nil
}
