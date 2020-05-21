package papertrail

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe a Papertrail logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetPapertrailInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Papertrail logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Papertrail logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	papertrail, err := c.Globals.Client.GetPapertrail(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", papertrail.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", papertrail.Version)
	fmt.Fprintf(out, "Name: %s\n", papertrail.Name)
	fmt.Fprintf(out, "Address: %s\n", papertrail.Address)
	fmt.Fprintf(out, "Port: %d\n", papertrail.Port)
	fmt.Fprintf(out, "Format: %s\n", papertrail.Format)
	fmt.Fprintf(out, "Format version: %d\n", papertrail.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", papertrail.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", papertrail.Placement)

	return nil
}
