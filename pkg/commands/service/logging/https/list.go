package https

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list HTTPS logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListHTTPSInput           `kong:"-"`
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

	o, err := c.Globals.APIClient.ListHTTPS(context.TODO(), &c.Input)
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
		for _, https := range o {
			tw.AddLine(
				fastly.ToValue(https.ServiceID),
				fastly.ToValue(https.ServiceVersion),
				fastly.ToValue(https.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, https := range o {
		fmt.Fprintf(out, "\tHTTPS %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(https.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(https.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(https.Name))
		fmt.Fprintf(out, "\t\tURL: %s\n", fastly.ToValue(https.URL))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(https.CompressionCodec))
		fmt.Fprintf(out, "\t\tContent type: %s\n", fastly.ToValue(https.ContentType))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(https.GzipLevel))
		fmt.Fprintf(out, "\t\tHeader name: %s\n", fastly.ToValue(https.HeaderName))
		fmt.Fprintf(out, "\t\tHeader value: %s\n", fastly.ToValue(https.HeaderValue))
		fmt.Fprintf(out, "\t\tMethod: %s\n", fastly.ToValue(https.Method))
		fmt.Fprintf(out, "\t\tJSON format: %s\n", fastly.ToValue(https.JSONFormat))
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", fastly.ToValue(https.TLSCACert))
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", fastly.ToValue(https.TLSClientCert))
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", fastly.ToValue(https.TLSClientKey))
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", fastly.ToValue(https.TLSHostname))
		fmt.Fprintf(out, "\t\tRequest max entries: %d\n", fastly.ToValue(https.RequestMaxEntries))
		fmt.Fprintf(out, "\t\tRequest max bytes: %d\n", fastly.ToValue(https.RequestMaxBytes))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(https.MessageType))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(https.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(https.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(https.ResponseCondition))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(https.Period))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(https.Placement))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(https.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
