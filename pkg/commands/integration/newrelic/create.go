package newrelic

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a New Relic notification integration.
type CreateCommand struct {
	argparser.Base

	// Required.
	IntegrationName string `name:"name" short:"n" required:"" help:"The name of the integration"`
	Account         string `name:"account-id" required:"" help:"The New Relic account ID"`
	APIKey          string `required:"" help:"The New Relic API key"`

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
		Config: map[string]string{"account": c.Account, "key": c.APIKey},
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

	text.Success(out, "Created New Relic integration '%s' (id: %s)", c.IntegrationName, fastly.ToValue(o.ID))
	return nil
}
