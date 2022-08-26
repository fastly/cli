package subscription

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

var states = []string{"pending", "processing", "issued", "renewing"}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS subscriptions")
	c.Globals = globals
	c.manifest = data

	// Optional Flags
	c.CmdClause.Flag("filter-active", "Limit the returned subscriptions to those that have currently active orders").BoolVar(&c.filterHasActiveOrder)
	c.CmdClause.Flag("filter-domain", "Limit the returned subscriptions to those that include the specific domain").StringVar(&c.filterTLSDomainID)
	c.CmdClause.Flag("filter-state", "Limit the returned subscriptions by state").HintOptions(states...).EnumVar(&c.filterState, states...)
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions(include...).EnumVar(&c.include, include...) // include is defined in ./describe.go
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

	filterHasActiveOrder bool
	filterState          string
	filterTLSDomainID    string
	include              string
	json                 bool
	manifest             manifest.Data
	pageNumber           int
	pageSize             int
	sort                 string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	rs, err := c.Globals.APIClient.ListTLSSubscriptions(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Filter Active":        c.filterHasActiveOrder,
			"Filter State":         c.filterState,
			"Filter TLS Domain ID": c.filterTLSDomainID,
			"Include":              c.include,
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
func (c *ListCommand) constructInput() *fastly.ListTLSSubscriptionsInput {
	var input fastly.ListTLSSubscriptionsInput

	if c.filterHasActiveOrder {
		input.FilterActiveOrders = c.filterHasActiveOrder
	}
	if c.filterState != "" {
		input.FilterState = c.filterState
	}
	if c.filterTLSDomainID != "" {
		input.FilterTLSDomainsID = c.filterTLSDomainID
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
	if c.sort != "" {
		input.Sort = c.sort
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
	if c.json {
		data, err := json.Marshal(rs)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

	t := text.NewTable(out)
	t.AddHeader("ID", "CERT AUTHORITY", "STATE", "CREATED")
	for _, r := range rs {
		t.AddLine(r.ID, r.CertificateAuthority, r.State, r.CreatedAt)
	}
	t.Print()
	return nil
}
