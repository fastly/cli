package alerts

import (
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("delete", "Delete Alert")

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying an Alert definition").Required().StringVar(&c.definitionID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete appropriate resource.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	definitionID string
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	err := c.Globals.APIClient.DeleteAlertDefinition(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Definition ID": c.definitionID,
		})
		return err
	}

	text.Success(out, "Deleted Alert entry '%s'", c.definitionID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteAlertDefinitionInput {
	input := fastly.DeleteAlertDefinitionInput{
		ID: &c.definitionID,
	}
	return &input
}
