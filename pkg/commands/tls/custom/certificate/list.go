package certificate

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

const emptyString = ""

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterNotAfter    string `help:"Limit the returned certificates to those that expire prior to the specified date in UTC"`
	FilterTLSDomainID string `name:"filter-domain" help:"Limit the returned certificates to those that include the specific domain"`
	Include           string `default:"" enum:"tls_activations," hints:"tls_activations" help:"Include related objects (comma-separated values)"`
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

	o, err := c.Globals.APIClient.ListCustomTLSCertificates(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter Not After":     c.FilterNotAfter,
			"Filter TLS Domain ID": c.FilterTLSDomainID,
			"Include":              c.Include,
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
func (c *ListCommand) constructInput() *fastly.ListCustomTLSCertificatesInput {
	var input fastly.ListCustomTLSCertificatesInput

	if c.FilterNotAfter != emptyString {
		input.FilterNotAfter = c.FilterNotAfter
	}
	if c.FilterTLSDomainID != emptyString {
		input.FilterTLSDomainsID = c.FilterTLSDomainID
	}
	if c.Include != emptyString {
		input.Include = c.Include
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
func printVerbose(out io.Writer, rs []*fastly.CustomTLSCertificate) {
	for _, r := range rs {
		fmt.Fprintf(out, "ID: %s\n", r.ID)
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

		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.CustomTLSCertificate) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "ISSUED TO", "NAME", "REPLACE", "SIGNATURE ALGORITHM")
	for _, r := range rs {
		t.AddLine(r.ID, r.IssuedTo, r.Name, r.Replace, r.SignatureAlgorithm)
	}
	t.Print()
	return nil
}
