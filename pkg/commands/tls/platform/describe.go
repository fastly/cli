package platform

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v6/fastly"
)

const include = "dns_records"

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Retrieve a single certificate").Alias("get")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS bulk certificate").Required().StringVar(&c.id)

	// Optional Flags
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	id       string
	json     bool
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	r, err := c.Globals.APIClient.GetBulkCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"TLS Bulk Certificate ID": c.id,
		})
		return err
	}

	err = c.print(out, r)
	if err != nil {
		return err
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetBulkCertificateInput {
	var input fastly.GetBulkCertificateInput

	input.ID = c.id

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.BulkCertificate) error {
	if c.json {
		data, err := json.Marshal(r)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

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
