package aclentry

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v8/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Retrieve a single ACL entry").Alias("get")

	// required
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Required().StringVar(&c.id)

	// optional
	c.RegisterFlagBool(c.JSONFlag()) // --json
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

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	aclID       string
	id          string
	manifest    manifest.Data
	serviceName cmd.OptionalServiceNameID
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
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
	input.ID = c.id
	input.ServiceID = serviceID

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, a *fastly.ACLEntry) error {
	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", a.ServiceID)
	}
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
	return nil
}
