package logentries

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

// ListCommand calls the Fastly API to list Logentries logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListLogentriesInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Logentries endpoints on a Fastly service version")
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

	logentriess, err := c.Globals.Client.ListLogentries(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, logentries := range logentriess {
			tw.AddLine(logentries.ServiceID, logentries.ServiceVersion, logentries.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, logentries := range logentriess {
		fmt.Fprintf(out, "\tLogentries %d/%d\n", i+1, len(logentriess))
		fmt.Fprintf(out, "\t\tService ID: %s\n", logentries.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", logentries.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", logentries.Name)
		fmt.Fprintf(out, "\t\tPort: %d\n", logentries.Port)
		fmt.Fprintf(out, "\t\tUse TLS: %t\n", logentries.UseTLS)
		fmt.Fprintf(out, "\t\tToken: %s\n", logentries.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", logentries.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", logentries.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", logentries.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", logentries.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
