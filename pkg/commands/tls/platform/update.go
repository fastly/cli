package platform

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command(
		"update", "Replace a certificate with a newly reissued certificate",
	)
	c.Globals = g

	// Required.

	c.CmdClause.Flag(
		"id", "Alphanumeric string identifying a TLS bulk certificate",
	).Required().StringVar(&c.id)

	c.CmdClause.Flag(
		"cert-blob", "The PEM-formatted certificate blob",
	).Required().StringVar(&c.certBlob)

	c.CmdClause.Flag(
		"intermediates-blob", "The PEM-formatted chain of intermediate blobs",
	).Required().StringVar(&c.intermediatesBlob)

	// Optional.

	c.CmdClause.Flag(
		"allow-untrusted", "Allow certificates that chain to untrusted roots",
	).Action(c.allowUntrusted.Set).BoolVar(&c.allowUntrusted.Value)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	allowUntrusted    argparser.OptionalBool
	certBlob          string
	id                string
	intermediatesBlob string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateBulkCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Bulk Certificate ID": c.id,
			"Allow Untrusted":         c.allowUntrusted.Value,
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

	input.ID = c.id
	input.CertBlob = c.certBlob
	input.IntermediatesBlob = c.intermediatesBlob

	if c.allowUntrusted.WasSet {
		input.AllowUntrusted = c.allowUntrusted.Value
	}

	return &input
}
