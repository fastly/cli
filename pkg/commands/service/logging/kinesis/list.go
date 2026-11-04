package kinesis

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Amazon Kinesis logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListKinesisInput         `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListKinesis(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, kinesis := range o {
			tw.AddLine(
				fastly.ToValue(kinesis.ServiceID),
				fastly.ToValue(kinesis.ServiceVersion),
				fastly.ToValue(kinesis.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kinesis := range o {
		fmt.Fprintf(out, "\tKinesis %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(kinesis.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(kinesis.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(kinesis.Name))
		fmt.Fprintf(out, "\t\tStream name: %s\n", fastly.ToValue(kinesis.StreamName))
		fmt.Fprintf(out, "\t\tRegion: %s\n", fastly.ToValue(kinesis.Region))
		if kinesis.AccessKey != nil || kinesis.SecretKey != nil {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(kinesis.AccessKey))
			fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(kinesis.SecretKey))
		}
		if kinesis.IAMRole != nil {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", fastly.ToValue(kinesis.IAMRole))
		}
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(kinesis.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(kinesis.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(kinesis.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(kinesis.Placement))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(kinesis.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
