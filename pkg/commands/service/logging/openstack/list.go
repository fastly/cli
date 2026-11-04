package openstack

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list OpenStack logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListOpenstackInput       `kong:"-"`
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

	o, err := c.Globals.APIClient.ListOpenstack(context.TODO(), &c.Input)
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
		for _, openstack := range o {
			tw.AddLine(
				fastly.ToValue(openstack.ServiceID),
				fastly.ToValue(openstack.ServiceVersion),
				fastly.ToValue(openstack.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, openstack := range o {
		fmt.Fprintf(out, "\tOpenstack %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(openstack.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(openstack.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(openstack.Name))
		fmt.Fprintf(out, "\t\tBucket: %s\n", fastly.ToValue(openstack.BucketName))
		fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(openstack.AccessKey))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(openstack.User))
		fmt.Fprintf(out, "\t\tURL: %s\n", fastly.ToValue(openstack.URL))
		fmt.Fprintf(out, "\t\tPath: %s\n", fastly.ToValue(openstack.Path))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(openstack.Period))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(openstack.GzipLevel))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(openstack.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(openstack.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(openstack.ResponseCondition))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(openstack.MessageType))
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", fastly.ToValue(openstack.TimestampFormat))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(openstack.Placement))
		fmt.Fprintf(out, "\t\tPublic key: %s\n", fastly.ToValue(openstack.PublicKey))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(openstack.CompressionCodec))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(openstack.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
