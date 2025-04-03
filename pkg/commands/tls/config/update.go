package config

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a TLS configuration")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS configuration").Required().StringVar(&c.id)
	c.CmdClause.Flag("name", "A custom name for your TLS configuration").Required().StringVar(&c.name)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	id   string
	name string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateCustomTLSConfiguration(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Configuration ID": c.id,
		})
		return err
	}

	text.Success(out, "Updated TLS Configuration '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateCustomTLSConfigurationInput {
	var input fastly.UpdateCustomTLSConfigurationInput

	input.ID = c.id
	input.Name = c.name

	return &input
}
