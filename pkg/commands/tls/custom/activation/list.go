package activation

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS activations")
	c.Globals = g

	// Optional.
	c.CmdClause.Flag("filter-cert", "Limit the returned activations to a specific certificate").StringVar(&c.filterTLSCertID)
	c.CmdClause.Flag("filter-config", "Limit the returned activations to a specific TLS configuration").StringVar(&c.filterTLSConfigID)
	c.CmdClause.Flag("filter-domain", "Limit the returned rules to a specific domain name").StringVar(&c.filterTLSDomainID)
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions(include...).EnumVar(&c.include, include...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	filterTLSCertID   string
	filterTLSConfigID string
	filterTLSDomainID string
	include           string
	pageNumber        int
	pageSize          int
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListTLSActivations(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter TLS Certificate ID":   c.filterTLSCertID,
			"Filter TLS Configuration ID": c.filterTLSConfigID,
			"Filter TLS Domain ID":        c.filterTLSDomainID,
			"Include":                     c.include,
			"Page Number":                 c.pageNumber,
			"Page Size":                   c.pageSize,
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
func (c *ListCommand) constructInput() *fastly.ListTLSActivationsInput {
	var input fastly.ListTLSActivationsInput

	if c.filterTLSCertID != "" {
		input.FilterTLSCertificateID = c.filterTLSCertID
	}
	if c.filterTLSConfigID != "" {
		input.FilterTLSConfigurationID = c.filterTLSConfigID
	}
	if c.filterTLSDomainID != "" {
		input.FilterTLSDomainID = c.filterTLSDomainID
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
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.TLSActivation) {
	for _, r := range rs {
		fmt.Fprintf(out, "\nID: %s\n", r.ID)

		if r.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
		}

		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.TLSActivation) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "CREATED_AT")
	for _, r := range rs {
		t.AddLine(r.ID, r.CreatedAt)
	}
	t.Print()
	return nil
}
