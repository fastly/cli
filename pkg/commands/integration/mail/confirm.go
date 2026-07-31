package mail

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ConfirmCommand calls the Fastly API to send a mailing list confirmation email.
type ConfirmCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	Address string
}

// NewConfirmCommand returns a usable command registered under the parent.
func NewConfirmCommand(parent argparser.Registerer, g *global.Data) *ConfirmCommand {
	c := ConfirmCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("confirm", "Send a mailing list confirmation email")

	// Required.
	c.CmdClause.Arg("address", "The mailing list address to send the confirmation email to").Required().StringVar(&c.Address)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ConfirmCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.CreateMailinglistConfirmationInput{Email: &c.Address}
	if err := c.Globals.APIClient.CreateMailinglistConfirmation(context.TODO(), input); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			Address string `json:"address"`
			Sent    bool   `json:"sent"`
		}{
			c.Address,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Sent confirmation email to '%s'", c.Address)
	return nil
}
