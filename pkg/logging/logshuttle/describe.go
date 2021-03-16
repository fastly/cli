package logshuttle

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Logshuttle logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetLogshuttleInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Logshuttle logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Logshuttle logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	logshuttle, err := c.Globals.Client.GetLogshuttle(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", logshuttle.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", logshuttle.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", logshuttle.Name)
	fmt.Fprintf(out, "URL: %s\n", logshuttle.URL)
	fmt.Fprintf(out, "Token: %s\n", logshuttle.Token)
	fmt.Fprintf(out, "Format: %s\n", logshuttle.Format)
	fmt.Fprintf(out, "Format version: %d\n", logshuttle.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", logshuttle.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", logshuttle.Placement)

	return nil
}
