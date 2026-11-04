package platform

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
	ID string `required:"" help:"Alphanumeric string identifying a TLS bulk certificate"`

	CertBlob          string `required:"" help:"The PEM-formatted certificate blob"`
	IntermediatesBlob string `required:"" help:"The PEM-formatted chain of intermediate blobs"`

	// Optional.
	AllowUntrusted argparser.OptionalBool `help:"Allow certificates that chain to untrusted roots"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateBulkCertificate(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Bulk Certificate ID": c.ID,
			"Allow Untrusted":         c.AllowUntrusted.Value,
		})
		return err
	}

	text.Success(out, "Updated TLS Bulk Certificate '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be
// used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateBulkCertificateInput {
	var input fastly.UpdateBulkCertificateInput

	input.ID = c.ID
	input.CertBlob = c.CertBlob
	input.IntermediatesBlob = c.IntermediatesBlob

	if c.AllowUntrusted.WasSet {
		input.AllowUntrusted = c.AllowUntrusted.Value
	}

	return &input
}
