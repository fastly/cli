package certificate

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
	c.CmdClause = parent.Command("create", "Create a TLS certificate").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob").Required().StringVar(&c.certBlob)

	// Optional flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").StringVar(&c.id)
	c.CmdClause.Flag("name", "A customizable name for your certificate. Defaults to the certificate's Common Name or first Subject Alternative Names (SAN) entry").StringVar(&c.name)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	certBlob string
	id       string
	manifest manifest.Data
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
