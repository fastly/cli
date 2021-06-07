package sumologic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Sumologic logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSumologicInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Sumologic logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Sumologic logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	// TODO(integralist): replace this surrounding code with cmd.ServiceDetails
	// once we have conditional boolean for the autoclone logic
	v, err := c.serviceVersion.Parse(c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.Input.ServiceVersion = v.Number

	sumologic, err := c.Globals.Client.GetSumologic(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", sumologic.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", sumologic.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", sumologic.Name)
	fmt.Fprintf(out, "URL: %s\n", sumologic.URL)
	fmt.Fprintf(out, "Format: %s\n", sumologic.Format)
	fmt.Fprintf(out, "Format version: %d\n", sumologic.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", sumologic.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", sumologic.MessageType)
	fmt.Fprintf(out, "Placement: %s\n", sumologic.Placement)

	return nil
}
