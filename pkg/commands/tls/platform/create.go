package platform

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to update an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	CertBlob          string `required:"" help:"The PEM-formatted certificate blob"`
	IntermediatesBlob string `required:"" help:"The PEM-formatted chain of intermediate blobs"`

	// Optional.
	AllowUntrusted argparser.OptionalBool `help:"Allow certificates that chain to untrusted roots"`
	Config         []string               `sep:"none" help:"Alphanumeric string identifying a TLS configuration (set flag once per Configuration ID)"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateBulkCertificate(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Allow Untrusted": c.AllowUntrusted.Value,
			"Configs":         c.Config,
		})
		return err
	}

	text.Success(out, "Uploaded TLS Bulk Certificate '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateBulkCertificateInput {
	var input fastly.CreateBulkCertificateInput

	input.CertBlob = c.CertBlob
	input.IntermediatesBlob = c.IntermediatesBlob

	if c.AllowUntrusted.WasSet {
		input.AllowUntrusted = c.AllowUntrusted.Value
	}

	var configs []*fastly.TLSConfiguration
	for _, v := range c.Config {
		configs = append(configs, &fastly.TLSConfiguration{ID: v})
	}
	input.Configurations = configs

	return &input
}
