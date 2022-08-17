package activations

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
	c.CmdClause = parent.Command("update", "Update the certificate used to terminate TLS traffic for the domain associated with this TLS activation")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("cert-id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.certID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS activation").Required().StringVar(&c.id)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	certID   string
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateTLSActivation(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
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
