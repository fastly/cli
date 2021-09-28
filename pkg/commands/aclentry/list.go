package aclentry

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List ACLs")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// Optional Flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	aclID    string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
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

	as, err := c.Globals.Client.ListACLEntries(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, as)
	} else {
		c.printSummary(out, as)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string) *fastly.ListACLEntriesInput {
	var input fastly.ListACLEntriesInput

	input.ACLID = c.aclID
	input.ServiceID = serviceID

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
func (c *ListCommand) printSummary(out io.Writer, as []*fastly.ACLEntry) {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "ID", "IP", "SUBNET", "NEGATED")
	for _, a := range as {
		t.AddLine(a.ServiceID, a.ID, a.IP, a.Subnet, a.Negated)
	}
	t.Print()
}
