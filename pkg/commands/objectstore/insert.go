package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// InsertKeyCommand calls the Fastly API to insert a key into an object store.
type InsertKeyCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.InsertObjectStoreKeyInput
}

// NewInsertKeyCommand returns a usable command registered under the parent.
func NewInsertKeyCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *InsertKeyCommand {
	var c InsertKeyCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("insert", "Insert key/value pair into a Fastly object store")
	c.CmdClause.Flag("id", "ID of Object Store").Short('n').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key", "Key to insert").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("value", "Value to insert").Required().StringVar(&c.Input.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *InsertKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	err := c.Globals.APIClient.InsertObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted key %s into object store %s", c.Input.Key, c.Input.ID)
	return nil
}
