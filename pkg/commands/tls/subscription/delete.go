package subscription

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Destroy a TLS subscription. A subscription cannot be destroyed if there are domains in the TLS enabled state").Alias("remove")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS subscription").Required().StringVar(&c.id)

	// Optional flags
	c.CmdClause.Flag("force", "A flag that allows you to edit and delete a subscription with active domains").Action(c.force.Set).BoolVar(&c.force.Value)

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	force    cmd.OptionalBool
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteTLSSubscription(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.id,
			"Force":               c.force.Value,
		})
		return err
	}

	text.Success(out, "Deleted TLS Subscription '%s' (force: %t)", c.id, c.force.Value)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteTLSSubscriptionInput {
	var input fastly.DeleteTLSSubscriptionInput

	input.ID = c.id

	if c.force.WasSet {
		input.Force = c.force.Value
	}

	return &input
}
