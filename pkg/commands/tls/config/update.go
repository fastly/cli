package config

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a TLS configuration")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS configuration").Required().StringVar(&c.id)
	c.CmdClause.Flag("name", "A custom name for your TLS configuration").Required().StringVar(&c.name)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	id       string
	manifest manifest.Data
	name     string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateCustomTLSConfiguration(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"TLS Configuration ID": c.id,
		})
		return err
	}

	text.Success(out, "Updated TLS Configuration '%s' (previously: '%s')", r.Name, input.Name)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateCustomTLSConfigurationInput {
	var input fastly.UpdateCustomTLSConfigurationInput

	input.ID = c.id
	input.Name = c.name

	return &input
}
