package platform

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
	c.CmdClause = parent.Command("describe", "Retrieve a single certificate").Alias("get")
	c.Globals = g
	c.manifest = m

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS bulk certificate").Required().StringVar(&c.id)

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

	o, err := c.Globals.APIClient.GetBulkCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Bulk Certificate ID": c.id,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetBulkCertificateInput {
	var input fastly.GetBulkCertificateInput

	input.ID = c.id

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.BulkCertificate) error {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)

	if r.NotAfter != nil {
		fmt.Fprintf(out, "Not after: %s\n", r.NotAfter)
	}
	if r.NotBefore != nil {
		fmt.Fprintf(out, "Not before: %s\n", r.NotBefore)
	}
	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
	}

	fmt.Fprintf(out, "Replace: %t\n", r.Replace)

	return nil
}
