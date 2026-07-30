package opsgenie

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an OpsGenie notification integration.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	ID     string
	APIKey string

	// Optional.
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
	c.CmdClause = parent.Command("update", "Update an OpsGenie notification integration")

	// Required.
	c.RegisterFlag(argparser.IntegrationIDFlag(&c.ID))
	c.CmdClause.Flag("api-key", "OpsGenie API key").Required().StringVar(&c.APIKey)

	// Optional.
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

	config := fastly.OpsGenieConfig{APIKey: c.APIKey}

	input := &fastly.UpdateIntegrationInput{
		ID:     c.ID,
		Type:   fastly.ToPointer(fastly.IntegrationTypeOpsGenie),
		Config: config.ToMap(),
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

	text.Success(out, "Updated OpsGenie integration (id: %s)", c.ID)
	return nil
}
