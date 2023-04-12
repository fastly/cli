package kvstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// CreateCommand calls the Fastly API to insert a key into an kv store.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.InsertKVStoreKeyInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Insert a key-value pair").Alias("insert")
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key-name", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("value", "Value").Required().StringVar(&c.Input.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.InsertKVStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted key %s into kv store %s", c.Input.Key, c.Input.ID)
	return nil
}
