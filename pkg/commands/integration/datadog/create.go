package datadog

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a Datadog notification integration.
type CreateCommand struct {
	argparser.Base

	// Required.
	IntegrationName string `name:"name" short:"n" required:"" help:"The name of the integration"`
	APIKey          string `required:"" help:"Datadog API key"`

	// Optional.
	Description argparser.OptionalString `help:"A description of the integration"`
	Site        argparser.OptionalString `help:"Datadog site, e.g. \"datadoghq.eu\" (defaults to the US site)"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	config := fastly.DatadogConfig{APIKey: c.APIKey}
	if c.Site.WasSet {
		config.Site = c.Site.Value
	}

	input := &fastly.CreateIntegrationInput{
		Name:   &c.IntegrationName,
		Type:   new(fastly.IntegrationTypeDatadog),
		Config: config.ToMap(),
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

	text.Success(out, "Created Datadog integration '%s' (id: %s)", c.IntegrationName, fastly.ToValue(o.ID))
	return nil
}
