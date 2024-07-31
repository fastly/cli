package aclentry

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Retrieve a single ACL entry").Alias("get")

	// Required.
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Required().StringVar(&c.id)

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
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	aclID       string
	id          string
	serviceName argparser.OptionalServiceNameID
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	o, err := c.Globals.APIClient.GetACLEntry(input)
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

	input.ACLID = c.aclID
	input.EntryID = c.id
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
