package cloudfiles

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a Cloudfiles logging endpoint.
type DescribeCommand struct {
	argparser.Base

	Input          fastly.GetCloudfilesInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"The name of the Cloudfiles logging object"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.GetCloudfiles(context.TODO(), &c.Input)
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

	lines := text.Lines{
		"Access key":         fastly.ToValue(o.AccessKey),
		"Bucket":             fastly.ToValue(o.BucketName),
		"Format version":     fastly.ToValue(o.FormatVersion),
		"Format":             fastly.ToValue(o.Format),
		"GZip level":         fastly.ToValue(o.GzipLevel),
		"Message type":       fastly.ToValue(o.MessageType),
		"Name":               fastly.ToValue(o.Name),
		"Path":               fastly.ToValue(o.Path),
		"Period":             fastly.ToValue(o.Period),
		"Placement":          fastly.ToValue(o.Placement),
		"Processing region":  fastly.ToValue(o.ProcessingRegion),
		"Public key":         fastly.ToValue(o.PublicKey),
		"Region":             fastly.ToValue(o.Region),
		"Response condition": fastly.ToValue(o.ResponseCondition),
		"Timestamp format":   fastly.ToValue(o.TimestampFormat),
		"User":               fastly.ToValue(o.User),
		"Version":            fastly.ToValue(o.ServiceVersion),
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(o.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
