package newrelic

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a New Relic notification integration.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	IntegrationName string
	Account         string
	APIKey          string

	// Optional.
	Description argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a New Relic notification integration").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the integration").Short('n').Required().StringVar(&c.IntegrationName)
	c.CmdClause.Flag("account-id", "The New Relic account ID").Required().StringVar(&c.Account)
	c.CmdClause.Flag("api-key", "The New Relic API key").Required().StringVar(&c.APIKey)

	// Optional.
	c.CmdClause.Flag("description", "A description of the integration").Action(c.Description.Set).StringVar(&c.Description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.CreateIntegrationInput{
		Name:   &c.IntegrationName,
		Type:   fastly.ToPointer(CommandName),
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
