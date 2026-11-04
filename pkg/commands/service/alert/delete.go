package alerts

import (
	"context"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DeleteCommand calls the Fastly API to delete appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required.
	DefinitionID string `name:"id" required:"" help:"Alphanumeric string identifying an Alert definition"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	err := c.Globals.APIClient.DeleteAlertDefinition(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Definition ID": c.DefinitionID,
		})
		return err
	}

	text.Success(out, "Deleted Alert entry '%s'", c.DefinitionID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteAlertDefinitionInput {
	input := fastly.DeleteAlertDefinitionInput{
		ID: &c.DefinitionID,
	}
	return &input
}
