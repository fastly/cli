package newrelic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Get the details of a New Relic Logs logging object for a particular service and version").Alias("get")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("name", "The name for the real-time logging configuration").Required().StringVar(&c.name)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional Flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	manifest       manifest.Data
	name           string
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	a, err := c.Globals.Client.GetNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	c.print(out, a)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) *fastly.GetNewRelicInput {
	var input fastly.GetNewRelicInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, l *fastly.NewRelic) {
	fmt.Fprintf(out, "\nService ID: %s\n", l.ServiceID)
	fmt.Fprintf(out, "Service Version: %d\n\n", l.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", l.Name)
	fmt.Fprintf(out, "Token: %s\n", l.Token)
	fmt.Fprintf(out, "Format: %s\n", l.Format)
	fmt.Fprintf(out, "Format Version: %d\n", l.FormatVersion)
	fmt.Fprintf(out, "Placement: %s\n", l.Placement)
	fmt.Fprintf(out, "Region: %s\n", l.Region)
	fmt.Fprintf(out, "Response Condition: %s\n\n", l.ResponseCondition)

	if l.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", l.CreatedAt)
	}
	if l.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", l.UpdatedAt)
	}
	if l.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", l.DeletedAt)
	}
}
