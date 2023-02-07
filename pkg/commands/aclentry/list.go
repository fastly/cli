package aclentry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("list", "List ACLs")

	// required
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// optional
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(cmd.PaginationDirection[0]).HintOptions(cmd.PaginationDirection...).EnumVar(&c.direction, cmd.PaginationDirection...)
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.perPage)
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.sort)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	aclID       string
	direction   string
	json        bool
	manifest    manifest.Data
	page        int
	perPage     int
	serviceName cmd.OptionalServiceNameID
	sort        string
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)
	paginator := c.Globals.APIClient.NewListACLEntriesPaginator(input)

	// TODO: Use generics support in go 1.18 to replace this almost identical
	// logic inside of 'dictionary-item list' and 'service list'.
	var as []*fastly.ACLEntry
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
		as = append(as, data...)
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, as)
	} else {
		err = c.printSummary(out, as)
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
	if c.json {
		data, err := json.Marshal(as)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

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
