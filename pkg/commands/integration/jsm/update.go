package jsm

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Jira Service Management notification integration.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	ID string

	// Optional.
	APIKey          argparser.OptionalString
	IntegrationName argparser.OptionalString
	Description     argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Jira Service Management notification integration")

	// Required.
	c.CmdClause.Arg("id", "Integration ID").Required().StringVar(&c.ID)

	// Optional.
	c.CmdClause.Flag("api-key", "Jira Service Management API key").Action(c.APIKey.Set).StringVar(&c.APIKey.Value)
	c.CmdClause.Flag("name", "The name of the integration").Short('n').Action(c.IntegrationName.Set).StringVar(&c.IntegrationName.Value)
	c.CmdClause.Flag("description", "A description of the integration").Action(c.Description.Set).StringVar(&c.Description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	config := map[string]string{}
	if c.APIKey.WasSet {
		config["apikey"] = c.APIKey.Value
	}

	input := &fastly.UpdateIntegrationInput{
		ID:   c.ID,
		Type: fastly.ToPointer(fastly.IntegrationTypeJSM),
	}
	if len(config) > 0 {
		input.Config = config
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

	text.Success(out, "Updated Jira Service Management integration (id: %s)", c.ID)
	return nil
}
