package domain

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

const emptyString = ""

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS domains")
	c.Globals = g

	// Optional.
	c.CmdClause.Flag("filter-cert", "Limit the returned domains to those listed in the given TLS certificate's SAN list").StringVar(&c.filterTLSCertsID)
	c.CmdClause.Flag("filter-in-use", "Limit the returned domains to those currently using Fastly to terminate TLS with SNI").Action(c.filterInUse.Set).BoolVar(&c.filterInUse.Value)
	c.CmdClause.Flag("filter-subscription", "Limit the returned domains to those for a given TLS subscription").StringVar(&c.filterTLSSubsID)
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

	filterInUse      argparser.OptionalBool
	filterTLSCertsID string
	filterTLSSubsID  string
	include          string
	pageNumber       int
	pageSize         int
	sort             string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListTLSDomains(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter In Use":            c.filterInUse,
			"Filter TLS Certificates":  c.filterTLSCertsID,
			"Filter TLS Subscriptions": c.filterTLSSubsID,
			"Include":                  c.include,
			"Page Number":              c.pageNumber,
			"Page Size":                c.pageSize,
			"Sort":                     c.sort,
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
func (c *ListCommand) constructInput() *fastly.ListTLSDomainsInput {
	var input fastly.ListTLSDomainsInput

	if c.filterInUse.WasSet {
		input.FilterInUse = &c.filterInUse.Value
	}
	if c.filterTLSCertsID != emptyString {
		input.FilterTLSCertificateID = c.filterTLSCertsID
	}
	if c.filterTLSSubsID != emptyString {
		input.FilterTLSSubscriptionID = c.filterTLSSubsID
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
func printVerbose(out io.Writer, rs []*fastly.TLSDomain) {
	for _, r := range rs {
		fmt.Fprintf(out, "\nID: %s\n", r.ID)
		fmt.Fprintf(out, "Type: %s\n", r.Type)
		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.TLSDomain) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "TYPE")
	for _, r := range rs {
		t.AddLine(r.ID, r.Type)
	}
	t.Print()
	return nil
}
