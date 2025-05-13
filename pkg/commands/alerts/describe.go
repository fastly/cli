package alerts

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("describe", "Describe Alert")

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying an Alert definition").Required().StringVar(&c.definitionID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	definitionID string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	definition, err := c.Globals.APIClient.GetAlertDefinition(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, definition); ok {
		return err
	}

	definitions := []*fastly.AlertDefinition{definition}
	if c.Globals.Verbose() {
		printVerbose(out, definitions)
	} else {
		printSummary(out, definitions)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetAlertDefinitionInput {
	input := fastly.GetAlertDefinitionInput{
		ID: &c.definitionID,
	}
	return &input
}
