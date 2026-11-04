package domain

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
	FilterTLSCertsID string                 `name:"filter-cert" help:"Limit the returned domains to those listed in the given TLS certificate's SAN list"`
	FilterInUse      argparser.OptionalBool `help:"Limit the returned domains to those currently using Fastly to terminate TLS with SNI"`
	FilterTLSSubsID  string                 `name:"filter-subscription" help:"Limit the returned domains to those for a given TLS subscription"`
	Include          string                 `default:"" enum:"tls_activations," hints:"tls_activations" help:"Include related objects (comma-separated values)"`
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

	o, err := c.Globals.APIClient.ListTLSDomains(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter In Use":            c.FilterInUse,
			"Filter TLS Certificates":  c.FilterTLSCertsID,
			"Filter TLS Subscriptions": c.FilterTLSSubsID,
			"Include":                  c.Include,
			"Page Number":              c.PageNumber,
			"Page Size":                c.PageSize,
			"Sort":                     c.Sort,
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

	if c.FilterInUse.WasSet {
		input.FilterInUse = &c.FilterInUse.Value
	}
	if c.FilterTLSCertsID != emptyString {
		input.FilterTLSCertificateID = c.FilterTLSCertsID
	}
	if c.FilterTLSSubsID != emptyString {
		input.FilterTLSSubscriptionID = c.FilterTLSSubsID
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
