package activation

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
	CertID string `required:"" help:"Alphanumeric string identifying a TLS certificate"`
	ID     string `required:"" help:"Alphanumeric string identifying a TLS activation"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateTLSActivation(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Activation ID":             c.ID,
			"TLS Activation Certificate ID": c.CertID,
		})
		return err
	}

	text.Success(out, "Updated TLS Activation Certificate '%s' (previously: '%s')", r.Certificate.ID, input.Certificate.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateTLSActivationInput {
	var input fastly.UpdateTLSActivationInput

	input.ID = c.ID
	input.Certificate = &fastly.CustomTLSCertificate{ID: c.CertID}

	return &input
}
