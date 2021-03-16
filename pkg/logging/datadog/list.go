package datadog

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Datadog logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListDatadogInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Datadog endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	datadogs, err := c.Globals.Client.ListDatadog(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, datadog := range datadogs {
			tw.AddLine(datadog.ServiceID, datadog.ServiceVersion, datadog.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, datadog := range datadogs {
		fmt.Fprintf(out, "\tDatadog %d/%d\n", i+1, len(datadogs))
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
