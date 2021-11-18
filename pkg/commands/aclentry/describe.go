package aclentry

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Retrieve a single ACL entry").Alias("get")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Required().StringVar(&c.id)

	// Optional Flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	aclID    string
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		err := errors.ErrNoServiceID
		c.Globals.ErrLog.Add(err)
		return err
	}

	input := c.constructInput(serviceID)

	a, err := c.Globals.Client.GetACLEntry(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	c.print(out, a)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string) *fastly.GetACLEntryInput {
	var input fastly.GetACLEntryInput

	input.ACLID = c.aclID
	input.ID = c.id
	input.ServiceID = serviceID

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, a *fastly.ACLEntry) {
	fmt.Fprintf(out, "\nService ID: %s\n", a.ServiceID)
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
}
