package aclentry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
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

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	aclID       string
	json        bool
	manifest    manifest.Data
	serviceName cmd.OptionalServiceNameID
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		var err error
		if !c.serviceName.WasSet {
			err = fsterr.ErrNoServiceID
			c.Globals.ErrLog.Add(err)
			return err
		}
		serviceID, err = c.serviceName.Parse(c.Globals.Client)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
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
func (c *ListCommand) printSummary(out io.Writer, as []*fastly.ACLEntry) error {
	if c.json {
		data, err := json.Marshal(as)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "ID", "IP", "SUBNET", "NEGATED")
	for _, a := range as {
		t.AddLine(a.ServiceID, a.ID, a.IP, a.Subnet, a.Negated)
	}
	t.Print()
	return nil
}
