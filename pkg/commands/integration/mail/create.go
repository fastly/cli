package mail

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a Mailing List notification integration.
type CreateCommand struct {
	argparser.Base

	// Required.
	IntegrationName string `name:"name" short:"n" required:"" help:"The name of the integration"`
	Address         string `required:"" help:"The mailing list address to notify"`

	// Optional.
	Description argparser.OptionalString `help:"A description of the integration"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.CreateIntegrationInput{
		Name:   &c.IntegrationName,
		Type:   new(CommandName),
		Config: map[string]string{"address": c.Address},
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	o, err := c.Globals.APIClient.CreateIntegration(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created Mailing List integration '%s' (id: %s)", c.IntegrationName, fastly.ToValue(o.ID))
	return nil
}
