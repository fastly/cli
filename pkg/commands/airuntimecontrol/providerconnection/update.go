package providerconnection

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/providerconnection"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a provider connection.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"Alphanumeric string identifying the provider connection"`

	// Optional.
	Models  []string                 `sep:"," help:"An updated comma-separated list of allowed AI model identifiers"`
	BaseURL argparser.OptionalString `help:"An updated base URL for the provider's API. Must be an absolute http or https URL"`
	APIKey  argparser.OptionalString `help:"An updated provider secret key for authentication"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &providerconnection.UpdateInput{
		ID: &c.ID,
	}
	if len(c.Models) > 0 {
		input.Models = c.Models
	}
	if c.BaseURL.WasSet {
		input.BaseURL = &c.BaseURL.Value
	}
	if c.APIKey.WasSet {
		input.APIKey = &c.APIKey.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := providerconnection.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintProviderConnection(out, data)
	return nil
}
