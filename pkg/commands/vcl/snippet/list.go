package snippet

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List the uploaded VCL snippets for a particular service and version")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

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

	json           bool
	manifest       manifest.Data
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	vs, err := c.Globals.APIClient.ListSnippets(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, serviceVersion.Number, vs)
	} else {
		err = c.printSummary(out, vs)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string, serviceVersion int) *fastly.ListSnippetsInput {
	var input fastly.ListSnippetsInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, serviceVersion int, vs []*fastly.Snippet) {
	fmt.Fprintf(out, "Service Version: %d\n", serviceVersion)

	for _, v := range vs {
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, "Name: %s\n", v.Name)
		fmt.Fprintf(out, "ID: %s\n", v.ID)
		fmt.Fprintf(out, "Priority: %d\n", v.Priority)
		fmt.Fprintf(out, "Dynamic: %t\n", cmd.IntToBool(v.Dynamic))
		fmt.Fprintf(out, "Type: %s\n", v.Type)
		fmt.Fprintf(out, "Content: \n%s\n", v.Content)

		if v.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", v.CreatedAt)
		}
		if v.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", v.UpdatedAt)
		}
		if v.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", v.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, ss []*fastly.Snippet) error {
	if c.json {
		data, err := json.Marshal(ss)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "VERSION", "NAME", "DYNAMIC", "SNIPPET ID")
	for _, s := range ss {
		t.AddLine(s.ServiceID, s.ServiceVersion, s.Name, cmd.IntToBool(s.Dynamic), s.ID)
	}
	t.Print()
	return nil
}
