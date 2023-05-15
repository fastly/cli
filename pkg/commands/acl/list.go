package acl

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
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

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
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

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	o, err := c.Globals.APIClient.ListACLs(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, serviceVersion.Number, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string, serviceVersion int) *fastly.ListACLsInput {
	var input fastly.ListACLsInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, serviceVersion int, as []*fastly.ACL) {
	fmt.Fprintf(out, "Service Version: %d\n\n", serviceVersion)

	for _, a := range as {
		fmt.Fprintf(out, "Name: %s\n", a.Name)
		fmt.Fprintf(out, "ID: %s\n\n", a.ID)

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
func (c *ListCommand) printSummary(out io.Writer, as []*fastly.ACL) error {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "VERSION", "NAME", "ID")
	for _, a := range as {
		t.AddLine(a.ServiceID, a.ServiceVersion, a.Name, a.ID)
	}
	t.Print()
	return nil
}
