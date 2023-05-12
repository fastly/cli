package kvstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an kv store.
type DeleteCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest manifest.Data
	Input    fastly.DeleteKVStoreInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("delete", "Delete an kv store")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := c.Globals.APIClient.DeleteKVStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.Input.ID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted KV Store %s", c.Input.ID)
	return nil
}
