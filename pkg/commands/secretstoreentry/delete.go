package secretstoreentry

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("delete", "Delete a secret")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name)) // --name
	c.RegisterFlag(cmd.StoreIDFlag(&c.Input.ID))  // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input fastly.DeleteSecretInput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := c.Globals.APIClient.DeleteSecret(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			Name    string `json:"name"`
			ID      string `json:"store_id"`
			Deleted bool   `json:"deleted"`
		}{
			c.Input.Name,
			c.Input.ID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted secret '%s' from Secret Store '%s'", c.Input.Name, c.Input.ID)
	return nil
}
