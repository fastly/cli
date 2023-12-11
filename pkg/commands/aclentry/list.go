package aclentry

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List ACLs")

	// Required.
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(argparser.PaginationDirection[0]).HintOptions(argparser.PaginationDirection...).EnumVar(&c.direction, argparser.PaginationDirection...)
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.perPage)
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.sort)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	aclID       string
	direction   string
	page        int
	perPage     int
	serviceName argparser.OptionalServiceNameID
	sort        string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)
	paginator := c.Globals.APIClient.NewListACLEntriesPaginator(input)

	// TODO: Use generics support in go 1.18 to replace this almost identical
	// logic inside of 'dictionary-item list' and 'service list'.
	var o []*fastly.ACLEntry
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"ACL ID":          c.aclID,
				"Service ID":      serviceID,
				"Remaining Pages": paginator.Remaining(),
			})
			return err
		}
		o = append(o, data...)
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
func (c *ListCommand) constructInput(serviceID string) *fastly.ListACLEntriesInput {
	var input fastly.ListACLEntriesInput

	input.ACLID = c.aclID
	input.Direction = c.direction
	input.Page = c.page
	input.PerPage = c.perPage
	input.ServiceID = serviceID
	input.Sort = c.sort

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, as []*fastly.ACLEntry) {
	for _, a := range as {
		fmt.Fprintf(out, "ACL ID: %s\n", a.ACLID)
		fmt.Fprintf(out, "ID: %s\n", a.ID)
		fmt.Fprintf(out, "IP: %s\n", a.IP)
		fmt.Fprintf(out, "Subnet: %d\n", a.Subnet)
		fmt.Fprintf(out, "Negated: %t\n", a.Negated)
		fmt.Fprintf(out, "Comment: %s\n\n", a.Comment)

		if a.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", a.CreatedAt)
		}
		if a.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", a.UpdatedAt)
		}
		if a.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", a.DeletedAt)
		}

		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, as []*fastly.ACLEntry) error {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "ID", "IP", "SUBNET", "NEGATED")
	for _, a := range as {
		var subnet int
		if a.Subnet != nil {
			subnet = *a.Subnet
		}
		t.AddLine(a.ServiceID, a.ID, a.IP, subnet, a.Negated)
	}
	t.Print()
	return nil
}
