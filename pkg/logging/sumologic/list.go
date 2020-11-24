package sumologic

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

// ListCommand calls the Fastly API to list Sumologic logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListSumologicsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Sumologic endpoints on a Fastly service version")
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

	sumologics, err := c.Globals.Client.ListSumologics(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, sumologic := range sumologics {
			tw.AddLine(sumologic.ServiceID, sumologic.ServiceVersion, sumologic.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, sumologic := range sumologics {
		fmt.Fprintf(out, "\tSumologic %d/%d\n", i+1, len(sumologics))
		fmt.Fprintf(out, "\t\tService ID: %s\n", sumologic.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", sumologic.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", sumologic.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", sumologic.URL)
		fmt.Fprintf(out, "\t\tFormat: %s\n", sumologic.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", sumologic.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", sumologic.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", sumologic.MessageType)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", sumologic.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
