package custom

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	NameFlag       string `name:"name" required:"" help:"The name of the VCL"`
	serviceVersion argparser.OptionalServiceVersion

	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	o, err := c.Globals.APIClient.GetVCL(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) *fastly.GetVCLInput {
	var input fastly.GetVCLInput

	input.Name = c.NameFlag
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, v *fastly.VCL) error {
	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(v.ServiceID))
	}
	fmt.Fprintf(out, "Service Version: %d\n\n", fastly.ToValue(v.ServiceVersion))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(v.Name))
	fmt.Fprintf(out, "Main: %t\n", fastly.ToValue(v.Main))
	fmt.Fprintf(out, "Content: \n%s\n\n", text.SanitizeTerminalOutput(fastly.ToValue(v.Content)))
	if v.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", v.CreatedAt)
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", v.UpdatedAt)
	}
	if v.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", v.DeletedAt)
	}
	return nil
}
