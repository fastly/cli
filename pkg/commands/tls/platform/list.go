package platform

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all certificates")
	c.Globals = globals
	c.manifest = data

	// Optional Flags
	c.CmdClause.Flag("filter-domain", "Optionally filter by the bulk attribute").StringVar(&c.filterTLSDomainID)
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)
	c.CmdClause.Flag("sort", "The order in which to list the results by creation date").StringVar(&c.sort)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	filterTLSDomainID string
	json              bool
	manifest          manifest.Data
	pageNumber        int
	pageSize          int
	sort              string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	rs, err := c.Globals.APIClient.ListBulkCertificates(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Filter TLS Domain ID": c.filterTLSDomainID,
			"Page Number":          c.pageNumber,
			"Page Size":            c.pageSize,
			"Sort":                 c.sort,
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
func (c *ListCommand) constructInput() *fastly.ListBulkCertificatesInput {
	var input fastly.ListBulkCertificatesInput

	if c.filterTLSDomainID != "" {
		input.FilterTLSDomainsIDMatch = c.filterTLSDomainID
	}
	if c.pageNumber > 0 {
		input.PageNumber = c.pageNumber
	}
	if c.pageSize > 0 {
		input.PageSize = c.pageSize
	}
	if c.sort != "" {
		input.Sort = c.sort
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.BulkCertificate) {
	for _, r := range rs {
		fmt.Fprintf(out, "ID: %s\n", r.ID)

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
		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.BulkCertificate) error {
	if c.json {
		data, err := json.Marshal(rs)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

	t := text.NewTable(out)
	t.AddHeader("ID", "REPLACE", "NOT BEFORE", "NOT AFTER", "CREATED")
	for _, r := range rs {
		t.AddLine(r.ID, r.Replace, r.NotBefore, r.NotAfter, r.CreatedAt)
	}
	t.Print()
	return nil
}
