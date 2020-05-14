package logshuttle

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list Logshuttle logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListLogshuttlesInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Logshuttle endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	logshuttles, err := c.Globals.Client.ListLogshuttles(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, logshuttle := range logshuttles {
			tw.AddLine(logshuttle.ServiceID, logshuttle.Version, logshuttle.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, logshuttle := range logshuttles {
		fmt.Fprintf(out, "\tLogshuttle %d/%d\n", i+1, len(logshuttles))
		fmt.Fprintf(out, "\t\tService ID: %s\n", logshuttle.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", logshuttle.Version)
		fmt.Fprintf(out, "\t\tName: %s\n", logshuttle.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", logshuttle.URL)
		fmt.Fprintf(out, "\t\tToken: %s\n", logshuttle.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", logshuttle.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", logshuttle.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", logshuttle.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", logshuttle.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
