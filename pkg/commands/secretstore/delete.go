package secretstore

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

	c.CmdClause = parent.Command("delete", "Delete a secret store")

	// Required.
	c.RegisterFlag(cmd.StoreIDFlag(&c.Input.ID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input    fastly.DeleteSecretStoreInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := cmd.Globals.APIClient.DeleteSecretStore(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			cmd.Input.ID,
			true,
		}
		_, err := cmd.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted secret store %s", cmd.Input.ID)

	return nil
}
