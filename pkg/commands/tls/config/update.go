package config

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID       string `required:"" help:"Alphanumeric string identifying a TLS configuration"`
	NameFlag string `name:"name" required:"" help:"A custom name for your TLS configuration"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateCustomTLSConfiguration(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Configuration ID": c.ID,
		})
		return err
	}

	text.Success(out, "Updated TLS Configuration '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateCustomTLSConfigurationInput {
	var input fastly.UpdateCustomTLSConfigurationInput

	input.ID = c.ID
	input.Name = c.NameFlag

	return &input
}
