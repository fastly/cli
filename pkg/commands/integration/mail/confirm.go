package mail

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ConfirmCommand calls the Fastly API to send a mailing list confirmation email.
type ConfirmCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput

	// Required.
	Address string `arg:"" required:"" help:"The mailing list address to send the confirmation email to"`
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
