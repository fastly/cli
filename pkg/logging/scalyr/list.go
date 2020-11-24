package scalyr

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// ListCommand calls the Fastly API to list Scalyr logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListScalyrsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Scalyr endpoints on a Fastly service version")
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

	scalyrs, err := c.Globals.Client.ListScalyrs(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, scalyr := range scalyrs {
			tw.AddLine(scalyr.ServiceID, scalyr.ServiceVersion, scalyr.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, scalyr := range scalyrs {
		fmt.Fprintf(out, "\tScalyr %d/%d\n", i+1, len(scalyrs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", scalyr.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", scalyr.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", scalyr.Name)
		fmt.Fprintf(out, "\t\tToken: %s\n", scalyr.Token)
		fmt.Fprintf(out, "\t\tRegion: %s\n", scalyr.Region)
		fmt.Fprintf(out, "\t\tFormat: %s\n", scalyr.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", scalyr.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", scalyr.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", scalyr.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
