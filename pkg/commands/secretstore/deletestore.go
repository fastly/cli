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

// NewDeleteStoreCommand returns a usable command registered under the parent.
func NewDeleteStoreCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteStoreCommand {
	c := DeleteStoreCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("delete", "Delete a secret store")

	// Required.
	c.RegisterFlag(storeIDFlag(&c.Input.ID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// DeleteStoreCommand calls the Fastly API to delete an appropriate resource.
type DeleteStoreCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.DeleteSecretStoreInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *DeleteStoreCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := cmd.Globals.APIClient.DeleteSecretStore(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.jsonOutput.enabled {
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
