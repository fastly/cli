package providerconnection

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/kingpin"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a provider connection.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	id string

	// Optional.
	models  []string
	baseURL argparser.OptionalString
	apiKey  argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a provider connection")

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying the provider connection").Required().StringVar(&c.id)

	// Optional.
	c.CmdClause.Flag("models", "An updated comma-separated list of allowed AI model identifiers").StringsVar(&c.models, kingpin.Separator(","))
	c.CmdClause.Flag("base-url", "An updated base URL for the provider's API").Action(c.baseURL.Set).StringVar(&c.baseURL.Value)
	c.CmdClause.Flag("api-key", "An updated provider secret key for authentication").Action(c.apiKey.Set).StringVar(&c.apiKey.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &providerconnection.UpdateInput{
		ID: &c.id,
	}
	if len(c.models) > 0 {
		input.Models = c.models
	}
	if c.baseURL.WasSet {
		input.BaseURL = &c.baseURL.Value
	}
	if c.apiKey.WasSet {
		input.APIKey = &c.apiKey.Value
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
