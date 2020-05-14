package honeycomb

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe a Honeycomb logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetHoneycombInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Honeycomb logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Honeycomb logging object").Short('d').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	honeycomb, err := c.Globals.Client.GetHoneycomb(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", honeycomb.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", honeycomb.Version)
	fmt.Fprintf(out, "Name: %s\n", honeycomb.Name)
	fmt.Fprintf(out, "Dataset: %s\n", honeycomb.Dataset)
	fmt.Fprintf(out, "Token: %s\n", honeycomb.Token)
	fmt.Fprintf(out, "Format: %s\n", honeycomb.Format)
	fmt.Fprintf(out, "Format version: %d\n", honeycomb.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", honeycomb.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", honeycomb.Placement)

	return nil
}
