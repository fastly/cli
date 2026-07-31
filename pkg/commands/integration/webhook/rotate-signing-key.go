package webhook

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RotateSigningKeyCommand calls the Fastly API to rotate the signing key for
// a Webhook notification integration.
type RotateSigningKeyCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	IntegrationID string
}

// NewRotateSigningKeyCommand returns a usable command registered under the parent.
func NewRotateSigningKeyCommand(parent argparser.Registerer, g *global.Data) *RotateSigningKeyCommand {
	c := RotateSigningKeyCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("rotate-signing-key", "Rotate the signing key for a Webhook notification integration")

	// Required.
	c.CmdClause.Arg("id", "Integration ID").Required().StringVar(&c.IntegrationID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *RotateSigningKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.RotateWebhookSigningKeyInput{IntegrationID: c.IntegrationID}

	o, err := c.Globals.APIClient.RotateWebhookSigningKey(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Signing key: '%s'", fastly.ToValue(o.SigningKey))
	return nil
}
