package certificate

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

const emptyString = ""

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS certificates")
	c.Globals = g

	// Optional.
	c.CmdClause.Flag("filter-not-after", "Limit the returned certificates to those that expire prior to the specified date in UTC").StringVar(&c.filterNotAfter)
	c.CmdClause.Flag("filter-domain", "Limit the returned certificates to those that include the specific domain").StringVar(&c.filterTLSDomainID)
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions("tls_activations").EnumVar(&c.include, "tls_activations")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)
	c.CmdClause.Flag("sort", "The order in which to list the results by creation date").StringVar(&c.sort)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	filterNotAfter    string
	filterTLSDomainID string
	include           string
	pageNumber        int
	pageSize          int
	sort              string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListCustomTLSCertificates(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter Not After":     c.filterNotAfter,
			"Filter TLS Domain ID": c.filterTLSDomainID,
			"Include":              c.include,
			"Page Number":          c.pageNumber,
			"Page Size":            c.pageSize,
			"Sort":                 c.sort,
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

	if c.filterNotAfter != emptyString {
		input.FilterNotAfter = c.filterNotAfter
	}
	if c.filterTLSDomainID != emptyString {
		input.FilterTLSDomainsID = c.filterTLSDomainID
	}
	if c.include != emptyString {
		input.Include = c.include
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
