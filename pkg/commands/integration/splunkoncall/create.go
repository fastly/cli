package splunkoncall

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a Splunk On-Call notification integration.
type CreateCommand struct {
	argparser.Base

	// Required.
	IntegrationName string `name:"name" short:"n" required:"" help:"The name of the integration"`
	URL             string `required:"" help:"The Splunk On-Call webhook URL"`

	// Optional.
	Description argparser.OptionalString `help:"A description of the integration"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	config := fastly.SplunkOnCallConfig{URL: c.URL}

	input := &fastly.CreateIntegrationInput{
		Name:   &c.IntegrationName,
		Type:   new(fastly.IntegrationTypeSplunkOnCall),
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

	text.Success(out, "Created Splunk On-Call integration '%s' (id: %s)", c.IntegrationName, fastly.ToValue(o.ID))
	return nil
}
