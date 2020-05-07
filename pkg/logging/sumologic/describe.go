package sumologic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe a Sumologic logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetSumologicInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Sumologic logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Sumologic logging object").Short('d').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	sumologic, err := c.Globals.Client.GetSumologic(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", sumologic.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", sumologic.Version)
	fmt.Fprintf(out, "Name: %s\n", sumologic.Name)
	fmt.Fprintf(out, "URL: %s\n", sumologic.URL)
	fmt.Fprintf(out, "Format: %s\n", sumologic.Format)
	fmt.Fprintf(out, "Format version: %d\n", sumologic.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", sumologic.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", sumologic.MessageType)
	fmt.Fprintf(out, "Placement: %s\n", sumologic.Placement)

	return nil
}
