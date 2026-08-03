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

// GetSigningKeyCommand calls the Fastly API to get the signing key for a
// Webhook notification integration.
type GetSigningKeyCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	IntegrationID string
}

// NewGetSigningKeyCommand returns a usable command registered under the parent.
func NewGetSigningKeyCommand(parent argparser.Registerer, g *global.Data) *GetSigningKeyCommand {
	c := GetSigningKeyCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get-signing-key", "Get the signing key for a Webhook notification integration")

	// Required.
	c.CmdClause.Arg("id", "Integration ID").Required().StringVar(&c.IntegrationID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetSigningKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.GetWebhookSigningKeyInput{IntegrationID: c.IntegrationID}

	o, err := c.Globals.APIClient.GetWebhookSigningKey(context.TODO(), input)
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
