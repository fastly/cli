package certificate

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Destroy a TLS certificate. TLS certificates already enabled for a domain cannot be destroyed").Alias("remove")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.id)

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteCustomTLSCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID": c.id,
		})
		return err
	}

	text.Success(out, "Deleted TLS Certificate '%s'", c.id)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteCustomTLSCertificateInput {
	var input fastly.DeleteCustomTLSCertificateInput

	input.ID = c.id

	return &input
}
