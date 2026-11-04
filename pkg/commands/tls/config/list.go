package config

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterBulk argparser.OptionalBool `help:"Optionally filter by the bulk attribute"`
	Include    string                 `default:"" enum:"dns_records," hints:"dns_records" help:"Include related objects (comma-separated values)"`
	argparser.JSONOutput
	PageNumber int `name:"page" help:"Page number of data set to fetch"`
	PageSize   int `name:"per-page" help:"Number of records per page"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListCustomTLSConfigurations(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter Bulk": c.FilterBulk,
			"Include":     c.Include,
			"Page Number": c.PageNumber,
			"Page Size":   c.PageSize,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomTLSConfigurationsInput {
	var input fastly.ListCustomTLSConfigurationsInput

	if c.FilterBulk.WasSet {
		input.FilterBulk = c.FilterBulk.Value
	}
	if c.Include != "" {
		input.Include = c.Include
	}
	if c.PageNumber > 0 {
		input.PageNumber = c.PageNumber
	}
	if c.PageSize > 0 {
		input.PageSize = c.PageSize
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.CustomTLSConfiguration) {
	for _, r := range rs {
		fmt.Fprintf(out, "ID: %s\n", r.ID)
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

		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.CustomTLSConfiguration) error {
	t := text.NewTable(out)
	t.AddHeader("NAME", "ID", "BULK", "DEFAULT", "TLS PROTOCOLS", "HTTP PROTOCOLS", "DNS RECORDS")
	for _, r := range rs {
		drs := make([]string, len(r.DNSRecords))
		for i, v := range r.DNSRecords {
			if v != nil {
				drs[i] = v.ID
			}
		}
		t.AddLine(
			r.Name,
			r.ID,
			r.Bulk,
			r.Default,
			strings.Join(r.TLSProtocols, ", "),
			strings.Join(r.HTTPProtocols, ", "),
			strings.Join(drs, ", "),
		)
	}
	t.Print()
	return nil
}
