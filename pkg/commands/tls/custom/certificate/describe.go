package certificate

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v8/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show a TLS certificate").Alias("get")
	c.Globals = g
	c.manifest = m

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.GetCustomTLSCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID": c.id,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetCustomTLSCertificateInput {
	var input fastly.GetCustomTLSCertificateInput

	input.ID = c.id

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.CustomTLSCertificate) error {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Issued to: %s\n", r.IssuedTo)
	fmt.Fprintf(out, "Issuer: %s\n", r.Issuer)
	fmt.Fprintf(out, "Name: %s\n", r.Name)

	if r.NotAfter != nil {
		fmt.Fprintf(out, "Not after: %s\n", r.NotAfter)
	}
	if r.NotBefore != nil {
		fmt.Fprintf(out, "Not before: %s\n", r.NotBefore)
	}

	fmt.Fprintf(out, "Replace: %t\n", r.Replace)
	fmt.Fprintf(out, "Serial number: %s\n", r.SerialNumber)
	fmt.Fprintf(out, "Signature algorithm: %s\n", r.SignatureAlgorithm)

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
	}

	return nil
}
