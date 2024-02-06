package activation

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
	c.CmdClause = parent.Command("enable", "Enable TLS for a particular TLS domain and certificate combination").Alias("add")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("cert-id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.certID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS activation").Required().StringVar(&c.id)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	certID string
	id     string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateTLSActivation(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Activation ID":             c.id,
			"TLS Activation Certificate ID": c.certID,
		})
		return err
	}

	text.Success(out, "Enabled TLS Activation '%s' (Certificate '%s')", r.ID, r.Certificate.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTLSActivationInput {
	var input fastly.CreateTLSActivationInput

	input.ID = c.id
	input.Certificate = &fastly.CustomTLSCertificate{ID: c.certID}

	return &input
}
