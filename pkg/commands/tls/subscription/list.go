package subscription

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

var states = []string{"pending", "processing", "issued", "renewing"}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterHasActiveOrder bool   `name:"filter-active" help:"Limit the returned subscriptions to those that have currently active orders"`
	FilterTLSDomainID    string `name:"filter-domain" help:"Limit the returned subscriptions to those that include the specific domain"`
	FilterState          string `default:"" enum:"pending,processing,issued,renewing," hints:"pending,processing,issued,renewing" help:"Limit the returned subscriptions by state"`
	Include              string `default:"" enum:"tls_authorizations,tls_authorizations.globalsign_email_challenge," hints:"tls_authorizations,tls_authorizations.globalsign_email_challenge" help:"Include related objects (comma-separated values)"`
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

	o, err := c.Globals.APIClient.ListTLSSubscriptions(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter Active":        c.FilterHasActiveOrder,
			"Filter State":         c.FilterState,
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
func (c *ListCommand) constructInput() *fastly.ListTLSSubscriptionsInput {
	var input fastly.ListTLSSubscriptionsInput

	if c.FilterHasActiveOrder {
		input.FilterActiveOrders = c.FilterHasActiveOrder
	}
	if c.FilterState != "" {
		input.FilterState = c.FilterState
	}
	if c.FilterTLSDomainID != "" {
		input.FilterTLSDomainsID = c.FilterTLSDomainID
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
	if c.Sort != "" {
		input.Sort = c.Sort
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.TLSSubscription) {
	for _, r := range rs {
		fmt.Fprintf(out, "ID: %s\n", r.ID)
		fmt.Fprintf(out, "Certificate Authority: %s\n", r.CertificateAuthority)
		fmt.Fprintf(out, "State: %s\n", r.State)

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
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.TLSSubscription) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "CERT AUTHORITY", "STATE", "CREATED")
	for _, r := range rs {
		t.AddLine(r.ID, r.CertificateAuthority, r.State, r.CreatedAt)
	}
	t.Print()
	return nil
}
