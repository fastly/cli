package aclentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Required.
	ACLID string `name:"acl-id" required:"" help:"Alphanumeric string identifying a ACL"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`

	Direction string `default:"ascend" enum:"ascend,descend" hints:"ascend,descend" help:"Direction in which to sort results"`
	Page      int    `help:"Page number of data set to fetch"`
	PerPage   int    `help:"Number of records per page"`
	Sort      string `default:"created" help:"Field on which to sort"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)
	paginator := c.Globals.APIClient.GetACLEntries(context.TODO(), input)

	var o []*fastly.ACLEntry
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"ACL ID":          c.ACLID,
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
func (c *ListCommand) constructInput(serviceID string) *fastly.GetACLEntriesInput {
	var input fastly.GetACLEntriesInput

	input.ACLID = c.ACLID
	if c.Direction != "" {
		input.Direction = new(c.Direction)
	}
	if c.Page > 0 {
		input.Page = new(c.Page)
	}
	if c.PerPage > 0 {
		input.PerPage = new(c.PerPage)
	}
	input.ServiceID = serviceID
	if c.Sort != "" {
		input.Sort = new(c.Sort)
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, as []*fastly.ACLEntry) {
	for _, a := range as {
		fmt.Fprintf(out, "ACL ID: %s\n", fastly.ToValue(a.ACLID))
		fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(a.EntryID))
		fmt.Fprintf(out, "IP: %s\n", fastly.ToValue(a.IP))
		fmt.Fprintf(out, "Subnet: %d\n", fastly.ToValue(a.Subnet))
		fmt.Fprintf(out, "Negated: %t\n", fastly.ToValue(a.Negated))
		fmt.Fprintf(out, "Comment: %s\n\n", fastly.ToValue(a.Comment))

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
		t.AddLine(
			fastly.ToValue(a.ServiceID),
			fastly.ToValue(a.EntryID),
			fastly.ToValue(a.IP),
			subnet,
			fastly.ToValue(a.Negated),
		)
	}
	t.Print()
	return nil
}
