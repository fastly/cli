package datadog

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Datadog logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetDatadogInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Datadog logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Datadog logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	datadog, err := c.Globals.Client.GetDatadog(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", datadog.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", datadog.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", datadog.Name)
	fmt.Fprintf(out, "Token: %s\n", datadog.Token)
	fmt.Fprintf(out, "Region: %s\n", datadog.Region)
	fmt.Fprintf(out, "Format: %s\n", datadog.Format)
	fmt.Fprintf(out, "Format version: %d\n", datadog.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", datadog.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", datadog.Placement)

	return nil
}
