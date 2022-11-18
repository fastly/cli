package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDeleteStoreCommand returns a usable command registered under the parent.
func NewDeleteStoreCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteStoreCommand {
	var c DeleteStoreCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("delete", "Delete secret store")
	c.RegisterFlag(storeIDFlag(&c.Input.ID))
	c.RegisterFlagBool(c.jsonFlag())
	return &c
}

// DeleteStoreCommand calls the Fastly API to delete a secret store.
type DeleteStoreCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteSecretStoreInput
	jsonOutput
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
