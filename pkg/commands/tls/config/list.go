package config

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS configurations")
	c.Globals = globals
	c.manifest = data

	// Optional Flags
	c.CmdClause.Flag("filter-bulk", "Optionally filter by the bulk attribute").Action(c.filterBulk.Set).BoolVar(&c.filterBulk.Value)
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions(include).EnumVar(&c.include, include)
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	filterBulk cmd.OptionalBool
	include    string
	json       bool
	manifest   manifest.Data
	pageNumber int
	pageSize   int
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	rs, err := c.Globals.APIClient.ListCustomTLSConfigurations(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter Bulk": c.filterBulk,
			"Include":     c.include,
			"Page Number": c.pageNumber,
			"Page Size":   c.pageSize,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, rs)
	} else {
		err = c.printSummary(out, rs)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomTLSConfigurationsInput {
	var input fastly.ListCustomTLSConfigurationsInput

	if c.filterBulk.WasSet {
		input.FilterBulk = c.filterBulk.Value
	}
	if c.include != "" {
		input.Include = c.include
	}
	if c.pageNumber > 0 {
		input.PageNumber = c.pageNumber
	}
	if c.pageSize > 0 {
		input.PageSize = c.pageSize
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
	if c.json {
		data, err := json.Marshal(rs)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

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
