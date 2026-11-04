package platform

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterTLSDomainID string `name:"filter-domain" help:"Optionally filter by the bulk attribute"`
	argparser.JSONOutput

	PageNumber int    `name:"page" help:"Page number of data set to fetch"`
	PageSize   int    `name:"per-page" help:"Number of records per page"`
	Sort       string `help:"The order in which to list the results by creation date"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListBulkCertificates(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter TLS Domain ID": c.FilterTLSDomainID,
			"Page Number":          c.PageNumber,
			"Page Size":            c.PageSize,
			"Sort":                 c.Sort,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListBulkCertificatesInput {
	var input fastly.ListBulkCertificatesInput

	if c.FilterTLSDomainID != "" {
		input.FilterTLSDomainsIDMatch = c.FilterTLSDomainID
	}
	if c.PageNumber > 0 {
		input.PageNumber = c.PageNumber
	}
	if c.PageSize > 0 {
		input.PageSize = c.PageSize
	}
	if c.Sort != "" {
		input.Sort = c.Sort
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, rs []*fastly.BulkCertificate) {
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
	t := text.NewTable(out)
	t.AddHeader("ID", "REPLACE", "NOT BEFORE", "NOT AFTER", "CREATED")
	for _, r := range rs {
		t.AddLine(r.ID, r.Replace, r.NotBefore, r.NotAfter, r.CreatedAt)
	}
	t.Print()
	return nil
}
