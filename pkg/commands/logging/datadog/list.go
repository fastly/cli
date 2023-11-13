package datadog

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Datadog logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.ListDatadogInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Datadog endpoints on a Fastly service version")

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
		Dst:         &g.Manifest.Flag.ServiceID,
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

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.ListDatadog(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, datadog := range o {
			tw.AddLine(datadog.ServiceID, datadog.ServiceVersion, datadog.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, datadog := range o {
		fmt.Fprintf(out, "\tDatadog %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", datadog.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", datadog.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", datadog.Name)
		fmt.Fprintf(out, "\t\tToken: %s\n", datadog.Token)
		fmt.Fprintf(out, "\t\tRegion: %s\n", datadog.Region)
		fmt.Fprintf(out, "\t\tFormat: %s\n", datadog.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", datadog.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", datadog.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", datadog.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
