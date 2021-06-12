package loggly

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Loggly logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListLogglyInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Loggly endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Envar(env.ServiceID).StringVar(&c.manifest.Flag.ServiceID)
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

	logglys, err := c.Globals.Client.ListLoggly(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, loggly := range logglys {
			tw.AddLine(loggly.ServiceID, loggly.ServiceVersion, loggly.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, loggly := range logglys {
		fmt.Fprintf(out, "\tLoggly %d/%d\n", i+1, len(logglys))
		fmt.Fprintf(out, "\t\tService ID: %s\n", loggly.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", loggly.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", loggly.Name)
		fmt.Fprintf(out, "\t\tToken: %s\n", loggly.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", loggly.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", loggly.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", loggly.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", loggly.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
