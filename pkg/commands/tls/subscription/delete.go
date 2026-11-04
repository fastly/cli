package subscription

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"Alphanumeric string identifying a TLS subscription"`

	// Optional.
	Force argparser.OptionalBool `help:"A flag that allows you to edit and delete a subscription with active domains"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteTLSSubscription(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.ID,
			"Force":               c.Force.Value,
		})
		return err
	}

	text.Success(out, "Deleted TLS Subscription '%s' (force: %t)", c.ID, c.Force.Value)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteTLSSubscriptionInput {
	var input fastly.DeleteTLSSubscriptionInput

	input.ID = c.ID

	if c.Force.WasSet {
		input.Force = c.Force.Value
	}

	return &input
}
