package papertrail

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

// ListCommand calls the Fastly API to list Papertrail logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListPapertrailsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Papertrail endpoints on a Fastly service version")
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

	papertrails, err := c.Globals.Client.ListPapertrails(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, papertrail := range papertrails {
			tw.AddLine(papertrail.ServiceID, papertrail.ServiceVersion, papertrail.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, papertrail := range papertrails {
		fmt.Fprintf(out, "\tPapertrail %d/%d\n", i+1, len(papertrails))
		fmt.Fprintf(out, "\t\tService ID: %s\n", papertrail.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", papertrail.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", papertrail.Name)
		fmt.Fprintf(out, "\t\tAddress: %s\n", papertrail.Address)
		fmt.Fprintf(out, "\t\tPort: %d\n", papertrail.Port)
		fmt.Fprintf(out, "\t\tFormat: %s\n", papertrail.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", papertrail.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", papertrail.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", papertrail.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
