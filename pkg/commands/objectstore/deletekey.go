package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// DeleteKeyCommand calls the Fastly API to delete an object store.
type DeleteKeyCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteObjectStoreKeyInput
}

// NewDeleteKeyCommand returns a usable command registered under the parent.
func NewDeleteKeyCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteKeyCommand {
	c := DeleteKeyCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("delete", "Delete a key")
	c.CmdClause.Flag("key-name", "Key name").Short('n').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.DeleteObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Deleted key %s from store ID %s", c.Input.Key, c.Input.ID)
	return nil
}
