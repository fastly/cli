package configstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("delete", "Delete a config store item")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "key",
		Short:       'k',
		Description: "Item name",
		Dst:         &c.input.Key,
		Required:    true,
	})
	c.RegisterFlag(cmd.StoreIDFlag(&c.input.StoreID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base
	cmd.JSONOutput

	input    fastly.DeleteConfigStoreItemInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := cmd.Globals.APIClient.DeleteConfigStoreItem(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.JSONOutput.Enabled {
		o := struct {
			StoreID string `json:"store_id"`
			Key     string `json:"item_key"`
			Deleted bool   `json:"deleted"`
		}{
			cmd.input.StoreID,
			cmd.input.Key,
			true,
		}
		_, err := cmd.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted config store item %s from store %s", cmd.input.Key, cmd.input.StoreID)

	return nil
}
