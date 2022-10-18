package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// GetKeyCommand calls the Fastly API to fetch the value of a key from an object store.
type GetKeyCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.GetObjectStoreKeyInput
}

// NewGetKeyCommand returns a usable command registered under the parent.
func NewGetKeyCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *GetKeyCommand {
	var c GetKeyCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("get", "Get Fastly edge config store key")
	c.CmdClause.Flag("id", "ID of object store").Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("k", "Key to fetch").Required().StringVar(&c.Input.Key)
	return &c
}

// Exec invokes the application logic for the command.
func (c *GetKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	value, err := c.Globals.APIClient.GetObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.PrintObjectStoreKeyValue(out, "", c.Input.Key, value)

	return nil
}
