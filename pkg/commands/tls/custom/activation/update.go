package activation

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
	c.CmdClause = parent.Command("update", "Update the certificate used to terminate TLS traffic for the domain associated with this TLS activation")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("cert-id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.certID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS activation").Required().StringVar(&c.id)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	certID string
	id     string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateTLSActivation(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Activation ID":             c.id,
			"TLS Activation Certificate ID": c.certID,
		})
		return err
	}

	text.Success(out, "Updated TLS Activation Certificate '%s' (previously: '%s')", r.Certificate.ID, input.Certificate.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateTLSActivationInput {
	var input fastly.UpdateTLSActivationInput

	input.ID = c.id
	input.Certificate = &fastly.CustomTLSCertificate{ID: c.certID}

	return &input
}
