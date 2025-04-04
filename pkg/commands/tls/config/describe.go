package config

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

const include = "dns_records"

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show a TLS configuration").Alias("get")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS configuration").Required().StringVar(&c.id)

	// Optional.
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions(include).EnumVar(&c.include, include)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	id      string
	include string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.GetCustomTLSConfiguration(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Configuration ID": c.id,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetCustomTLSConfigurationInput {
	var input fastly.GetCustomTLSConfigurationInput

	input.ID = c.id

	if c.include != "" {
		input.Include = c.include
	}

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.CustomTLSConfiguration) error {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Name: %s\n", r.Name)

	if len(r.DNSRecords) > 0 {
		for _, v := range r.DNSRecords {
			if v != nil {
				fmt.Fprintf(out, "DNS Record ID: %s\n", v.ID)
				fmt.Fprintf(out, "DNS Record Type: %s\n", v.RecordType)
				fmt.Fprintf(out, "DNS Record Region: %s\n", v.Region)
			}
		}
	}

	fmt.Fprintf(out, "Bulk: %t\n", r.Bulk)
	fmt.Fprintf(out, "Default: %t\n", r.Default)

	if len(r.HTTPProtocols) > 0 {
		for _, v := range r.HTTPProtocols {
			fmt.Fprintf(out, "HTTP Protocol: %s\n", v)
		}
	}

	if len(r.TLSProtocols) > 0 {
		for _, v := range r.TLSProtocols {
			fmt.Fprintf(out, "TLS Protocol: %s\n", v)
		}
	}

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
	}

	return nil
}
