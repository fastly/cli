package aclentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	ACLID string `name:"acl-id" required:"" help:"Alphanumeric string identifying a ACL"`
	ID    string `required:"" help:"Alphanumeric string identifying an ACL Entry"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	o, err := c.Globals.APIClient.GetACLEntry(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string) *fastly.GetACLEntryInput {
	var input fastly.GetACLEntryInput

	input.ACLID = c.ACLID
	input.EntryID = c.ID
	input.ServiceID = serviceID

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, a *fastly.ACLEntry) error {
	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(a.ServiceID))
	}
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
	return nil
}
