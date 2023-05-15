package platform

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Destroy a certificate. This disables TLS for all domains listed as SAN entries").Alias("remove")
	c.Globals = g
	c.manifest = m

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS bulk certificate").Required().StringVar(&c.id)

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteBulkCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Bulk Certificate ID": c.id,
		})
		return err
	}

	text.Success(out, "Deleted TLS Bulk Certificate '%s'", c.id)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteBulkCertificateInput {
	var input fastly.DeleteBulkCertificateInput

	input.ID = c.id

	return &input
}
