package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// KeysCommand calls the Fastly API to list the keys for a given object store.
type KeysCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListObjectStoreKeysInput
}

// NewKeysCommand returns a usable command registered under the parent.
func NewKeysCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *KeysCommand {
	var c KeysCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("keys", "List Fastly object store keys")
	c.CmdClause.Flag("id", "ID of object store").Required().StringVar(&c.Input.ID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *KeysCommand) Exec(_ io.Reader, out io.Writer) error {
	o, err := c.Globals.APIClient.ListObjectStoreKeys(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.PrintObjectStoreKeys(out, "", o.Data)

	return nil
}
