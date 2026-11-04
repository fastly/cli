package condition

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput `kong:"-"`

	serviceVersion argparser.OptionalServiceVersion

	// Required flags
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional Flags
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
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

	var input fastly.ListConditionsInput

	input.ServiceID = serviceID
	input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListConditions(context.TODO(), &input)
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

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME", "STATEMENT", "TYPE", "PRIORITY")
		for _, r := range o {
			tw.AddLine(
				fastly.ToValue(r.ServiceID),
				fastly.ToValue(r.ServiceVersion),
				fastly.ToValue(r.Name),
				fastly.ToValue(r.Statement),
				fastly.ToValue(r.Type),
				fastly.ToValue(r.Priority),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", input.ServiceVersion)
	for i, r := range o {
		fmt.Fprintf(out, "\tCondition %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(r.Name))
		fmt.Fprintf(out, "\t\tStatement: %v\n", fastly.ToValue(r.Statement))
		fmt.Fprintf(out, "\t\tType: %v\n", fastly.ToValue(r.Type))
		fmt.Fprintf(out, "\t\tPriority: %v\n", fastly.ToValue(r.Priority))
	}
	fmt.Fprintln(out)

	return nil
}
