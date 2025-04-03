package activation

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("disable", "Disable TLS on the domain associated with this TLS activation").Alias("remove")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS activation").Required().StringVar(&c.id)

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	id string
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteTLSActivation(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Activation ID": c.id,
		})
		return err
	}

	text.Success(out, "Disabled TLS Activation '%s'", c.id)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteTLSActivationInput {
	var input fastly.DeleteTLSActivationInput

	input.ID = c.id

	return &input
}
