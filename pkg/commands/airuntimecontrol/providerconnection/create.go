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

// CreateCommand calls the Fastly API to create a provider connection.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string   `name:"name" required:"" help:"Human-readable name of the provider connection"`
	Models   []string `required:"" sep:"," help:"Comma-separated list of allowed AI model identifiers"`
	BaseURL  string   `required:"" help:"Base URL for the provider's API. Must be an absolute http or https URL"`
	APIKey   string   `required:"" help:"Provider's secret key for authentication"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := providerconnection.Create(context.TODO(), fc, &providerconnection.CreateInput{
		Name:    &c.NameFlag,
		Models:  c.Models,
		BaseURL: &c.BaseURL,
		APIKey:  &c.APIKey,
	})
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
