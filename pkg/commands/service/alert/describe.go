package alerts

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	DefinitionID string `name:"id" required:"" help:"Alphanumeric string identifying an Alert definition"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	definition, err := c.Globals.APIClient.GetAlertDefinition(context.TODO(), input)
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
		ID: &c.DefinitionID,
	}
	return &input
}
