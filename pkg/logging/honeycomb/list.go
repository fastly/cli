package honeycomb

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Honeycomb logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListHoneycombsInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Honeycomb endpoints on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	honeycombs, err := c.Globals.Client.ListHoneycombs(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, honeycomb := range honeycombs {
			tw.AddLine(honeycomb.ServiceID, honeycomb.ServiceVersion, honeycomb.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
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
