package platform

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("upload", "Upload a new certificate")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob").Required().StringVar(&c.certBlob)
	c.CmdClause.Flag("intermediates-blob", "The PEM-formatted chain of intermediate blobs").Required().StringVar(&c.intermediatesBlob)

	// Optional.
	c.CmdClause.Flag("allow-untrusted", "Allow certificates that chain to untrusted roots").Action(c.allowUntrusted.Set).BoolVar(&c.allowUntrusted.Value)
	c.CmdClause.Flag("config", "Alphanumeric string identifying a TLS configuration (set flag once per Configuration ID)").StringsVar(&c.config)

	return &c
}

// CreateCommand calls the Fastly API to update an appropriate resource.
type CreateCommand struct {
	argparser.Base

	allowUntrusted    argparser.OptionalBool
	certBlob          string
	config            []string
	intermediatesBlob string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateBulkCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Allow Untrusted": c.allowUntrusted.Value,
			"Configs":         c.config,
		})
		return err
	}

	text.Success(out, "Uploaded TLS Bulk Certificate '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateBulkCertificateInput {
	var input fastly.CreateBulkCertificateInput

	input.CertBlob = c.certBlob
	input.IntermediatesBlob = c.intermediatesBlob

	if c.allowUntrusted.WasSet {
		input.AllowUntrusted = c.allowUntrusted.Value
	}

	var configs []*fastly.TLSConfiguration
	for _, v := range c.config {
		configs = append(configs, &fastly.TLSConfiguration{ID: v})
	}
	input.Configurations = configs

	return &input
}
