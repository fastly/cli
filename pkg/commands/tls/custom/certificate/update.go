package certificate

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Replace a TLS certificate with a newly reissued TLS certificate, or update a TLS certificate's name")
	c.Globals = globals
	c.manifest = data

	// required
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob").Required().StringVar(&c.certBlob)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.id)

	// optional
	c.CmdClause.Flag("name", "A customizable name for your certificate. Defaults to the certificate's Common Name or first Subject Alternative Names (SAN) entry").StringVar(&c.name)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	certBlob string
	id       string
	manifest manifest.Data
	name     string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateCustomTLSCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID":   c.id,
			"TLS Certificate Name": c.name,
		})
		return err
	}

	if c.name != "" {
		text.Success(out, "Updated TLS Certificate '%s' (previously: '%s')", r.Name, input.Name)
	} else {
		text.Success(out, "Updated TLS Certificate '%s'", r.ID)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateCustomTLSCertificateInput {
	var input fastly.UpdateCustomTLSCertificateInput

	input.ID = c.id
	input.CertBlob = c.certBlob

	if c.name != "" {
		input.Name = c.name
	}

	return &input
}
