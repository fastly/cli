package provider

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/provider"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListModelsCommand calls the Fastly API to list the models available for a
// specific provider.
type ListModelsCommand struct {
	argparser.Base

	// Required.
	ProviderID string `required:"" help:"Alphanumeric string identifying the provider (e.g. \"anthropic\", \"openai\")"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListModelsCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := provider.ListModels(context.TODO(), fc, &provider.ListModelsInput{
		ProviderID: &c.ProviderID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintModelsTbl(out, data.Data)
	return nil
}
