package provider

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/provider"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListModelsCommand calls the Fastly API to list the models available for a
// specific provider.
type ListModelsCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	providerID string
}

// NewListModelsCommand returns a usable command registered under the parent.
func NewListModelsCommand(parent argparser.Registerer, g *global.Data) *ListModelsCommand {
	c := ListModelsCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list-models", "List models available for a specific provider")

	// Required.
	c.CmdClause.Flag("provider-id", "Alphanumeric string identifying the provider (e.g. \"anthropic\", \"openai\")").Required().StringVar(&c.providerID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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
		ProviderID: &c.providerID,
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
