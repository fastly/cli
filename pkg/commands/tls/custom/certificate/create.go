package certificate

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a TLS certificate").Alias("add")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob").Required().StringVar(&c.certBlob)

	// Optional.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").StringVar(&c.id)
	c.CmdClause.Flag("name", "A customizable name for your certificate. Defaults to the certificate's Common Name or first Subject Alternative Names (SAN) entry").StringVar(&c.name)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	certBlob string
	id       string
	name     string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateCustomTLSCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID":   c.id,
			"TLS Certificate Name": c.name,
		})
		return err
	}

	text.Success(out, "Created TLS Certificate '%s'", r.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateCustomTLSCertificateInput {
	var input fastly.CreateCustomTLSCertificateInput

	if c.id != "" {
		input.ID = c.id
	}
	input.CertBlob = c.certBlob
	if c.name != "" {
		input.Name = c.name
	}

	return &input
}
