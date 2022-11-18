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

// NewDeleteSecretCommand returns a usable command registered under the parent.
func NewDeleteSecretCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteSecretCommand {
	var c DeleteSecretCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("delete", "Delete secret")
	c.RegisterFlag(storeIDFlag(&c.Input.ID))
	c.RegisterFlag(secretNameFlag(&c.Input.Name))
	c.RegisterFlagBool(c.jsonFlag())
	return &c
}

// DeleteSecretCommand calls the Fastly API to delete a secret.
type DeleteSecretCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteSecretInput
	jsonOutput
}

// Exec invokes the application logic for the command.
func (cmd *DeleteSecretCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := cmd.Globals.APIClient.DeleteSecret(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.jsonOutput.enabled {
		o := struct {
			Name    string `json:"name"`
			ID      string `json:"store_id"`
			Deleted bool   `json:"deleted"`
		}{
			cmd.Input.Name,
			cmd.Input.ID,
			true,
		}
		_, err := cmd.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted secret %s from store %s", cmd.Input.Name, cmd.Input.ID)

	return nil
}
