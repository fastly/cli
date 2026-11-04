package condition

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required flags
	NameFlag       string `name:"name" short:"n" required:"" help:"Name of condition"`
	serviceVersion argparser.OptionalServiceVersion

	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional flags
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return errors.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	var input fastly.GetConditionInput
	input.ServiceID = serviceID
	input.ServiceVersion = fastly.ToValue(serviceVersion.Number)
	input.Name = c.NameFlag

	r, err := c.Globals.APIClient.GetCondition(context.TODO(), &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, r); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(r.ServiceID))
	}
	fmt.Fprintf(out, "Version: %d\n", fastly.ToValue(r.ServiceVersion))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(r.Name))
	fmt.Fprintf(out, "Statement: %s\n", fastly.ToValue(r.Statement))
	fmt.Fprintf(out, "Type: %s\n", fastly.ToValue(r.Type))
	fmt.Fprintf(out, "Priority: %d\n", fastly.ToValue(r.Priority))

	return nil
}
