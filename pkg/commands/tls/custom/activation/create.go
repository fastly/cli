package activation

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	CertID      string `required:"" help:"Alphanumeric string identifying a TLS certificate"`
	TLSConfigID string `required:"" help:"Alphanumeric string identifying a TLS configuration"`
	TLSDomain   string `required:"" help:"The domain name associated with the TLS certificate"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateTLSActivation(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Configuration ID":          c.TLSConfigID,
			"TLS Activation Certificate ID": c.CertID,
			"TLS Domain":                    c.TLSDomain,
		})
		return err
	}

	text.Success(out, "Enabled TLS Activation '%s' (Certificate '%s', Configuration '%s')", r.ID, c.CertID, c.TLSConfigID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTLSActivationInput {
	var input fastly.CreateTLSActivationInput

	input.Configuration = &fastly.TLSConfiguration{ID: c.TLSConfigID}
	input.Certificate = &fastly.CustomTLSCertificate{ID: c.CertID}
	input.Domain = &fastly.TLSDomain{ID: c.TLSDomain}

	return &input
}
