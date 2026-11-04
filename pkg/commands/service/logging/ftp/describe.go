package ftp

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe an FTP logging endpoint.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.GetFTPInput              `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion

	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
	NameFlag string `name:"name" short:"n" required:"" help:"The name of the FTP logging object"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	c.Input.Name = c.NameFlag
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

	o, err := c.Globals.APIClient.GetFTP(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Address":            fastly.ToValue(o.Address),
		"Compression codec":  fastly.ToValue(o.CompressionCodec),
		"Format version":     fastly.ToValue(o.FormatVersion),
		"Format":             fastly.ToValue(o.Format),
		"GZip level":         fastly.ToValue(o.GzipLevel),
		"Name":               fastly.ToValue(o.Name),
		"Password":           fastly.ToValue(o.Password),
		"Path":               fastly.ToValue(o.Path),
		"Period":             fastly.ToValue(o.Period),
		"Placement":          fastly.ToValue(o.Placement),
		"Port":               fastly.ToValue(o.Port),
		"Processing region":  fastly.ToValue(o.ProcessingRegion),
		"Public key":         fastly.ToValue(o.PublicKey),
		"Response condition": fastly.ToValue(o.ResponseCondition),
		"Timestamp format":   fastly.ToValue(o.TimestampFormat),
		"Username":           fastly.ToValue(o.Username),
		"Version":            fastly.ToValue(o.ServiceVersion),
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(o.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
