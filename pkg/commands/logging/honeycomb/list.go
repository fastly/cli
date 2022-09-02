package honeycomb

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

// ListCommand calls the Fastly API to list Honeycomb logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListHoneycombsInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Honeycomb endpoints on a Fastly service version")
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
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	honeycombs, err := c.Globals.APIClient.ListHoneycombs(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(honeycombs)
			if err != nil {
				return err
			}
			out.Write(data)
			return nil
		}

		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, honeycomb := range honeycombs {
			tw.AddLine(honeycomb.ServiceID, honeycomb.ServiceVersion, honeycomb.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, honeycomb := range honeycombs {
		fmt.Fprintf(out, "\tHoneycomb %d/%d\n", i+1, len(honeycombs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", honeycomb.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", honeycomb.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", honeycomb.Name)
		fmt.Fprintf(out, "\t\tDataset: %s\n", honeycomb.Dataset)
		fmt.Fprintf(out, "\t\tToken: %s\n", honeycomb.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", honeycomb.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", honeycomb.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", honeycomb.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", honeycomb.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
