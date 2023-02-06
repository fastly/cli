package objectstorekeys

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// InsertCommand calls the Fastly API to insert a key into an object store.
type InsertCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.InsertObjectStoreKeyInput
}

// NewInsertCommand returns a usable command registered under the parent.
func NewInsertCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *InsertCommand {
	c := InsertCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("insert", "Insert a key-value pair")
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key-name", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("value", "Value").Required().StringVar(&c.Input.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *InsertCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.InsertObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted key %s into object store %s", c.Input.Key, c.Input.ID)
	return nil
}
