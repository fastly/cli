package platform

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("upload", "Upload a new certificate")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob").Required().StringVar(&c.certBlob)
	c.CmdClause.Flag("intermediates-blob", "The PEM-formatted chain of intermediate blobs").Required().StringVar(&c.intermediatesBlob)

	// Optional flags
	c.CmdClause.Flag("allow-untrusted", "Allow certificates that chain to untrusted roots").Action(c.allowUntrusted.Set).BoolVar(&c.allowUntrusted.Value)
	c.CmdClause.Flag("config", "Alphanumeric string identifying a TLS configuration (set flag once per Configuration ID)").StringsVar(&c.config)

	return &c
}

// CreateCommand calls the Fastly API to update an appropriate resource.
type CreateCommand struct {
	cmd.Base

	allowUntrusted    cmd.OptionalBool
	certBlob          string
	config            []string
	intermediatesBlob string
	manifest          manifest.Data
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
