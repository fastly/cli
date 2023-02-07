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

// NewDeleteSecretCommand returns a usable command registered under the parent.
func NewDeleteSecretCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteSecretCommand {
	c := DeleteSecretCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("delete", "Delete a secret")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name)) // --name
	c.RegisterFlag(storeIDFlag(&c.Input.ID))      // --store-id

	// Optional.
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// DeleteSecretCommand calls the Fastly API to delete an appropriate resource.
type DeleteSecretCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.DeleteSecretInput
	manifest manifest.Data
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
