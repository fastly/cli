package newrelic

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a New Relic notification integration.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID      string `arg:"" required:"" help:"Integration ID"`
	Account string `name:"account-id" required:"" help:"The New Relic account ID"`
	APIKey  string `required:"" help:"The New Relic API key"`

	// Optional.
	IntegrationName argparser.OptionalString `name:"name" short:"n" help:"The name of the integration"`
	Description     argparser.OptionalString `help:"A description of the integration"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.UpdateIntegrationInput{
		ID:     c.ID,
		Type:   new(CommandName),
		Config: map[string]string{"account": c.Account, "key": c.APIKey},
	}
	if c.IntegrationName.WasSet {
		input.Name = &c.IntegrationName.Value
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	if err := c.Globals.APIClient.UpdateIntegration(context.TODO(), input); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Updated bool   `json:"updated"`
		}{
			c.ID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Updated New Relic integration (id: %s)", c.ID)
	return nil
}
