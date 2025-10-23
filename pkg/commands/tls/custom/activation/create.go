package activation

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

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
	c.CmdClause.Flag("tls-config-id", "Alphanumeric string identifying a TLS configuration").Required().StringVar(&c.tlsConfigId)
	c.CmdClause.Flag("tls-domain", "The domain name associated with the TLS activation").Required().StringVar(&c.tlsDomain)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	certID      string
	tlsConfigId string
	tlsDomain   string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateTLSActivation(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Configuration ID":          c.tlsConfigId,
			"TLS Activation Certificate ID": c.certID,
			"TLS Domain":                    c.tlsDomain,
		})
		return err
	}

	text.Success(out, "Enabled TLS Activation '%s' (Certificate '%s', Configuration '%s')", r.ID, c.certID, c.tlsConfigId)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTLSActivationInput {
	var input fastly.CreateTLSActivationInput

	input.Configuration = &fastly.TLSConfiguration{ID: c.tlsConfigId}
	input.Certificate = &fastly.CustomTLSCertificate{ID: c.certID}
	input.Domain = &fastly.TLSDomain{ID: c.tlsDomain}

	return &input
}
