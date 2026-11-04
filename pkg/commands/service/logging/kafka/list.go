package kafka

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Kafka logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListKafkasInput          `kong:"-"`
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

	o, err := c.Globals.APIClient.ListKafkas(context.TODO(), &c.Input)
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
		for _, kafka := range o {
			tw.AddLine(
				fastly.ToValue(kafka.ServiceID),
				fastly.ToValue(kafka.ServiceVersion),
				fastly.ToValue(kafka.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kafka := range o {
		fmt.Fprintf(out, "\tKafka %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(kafka.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(kafka.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(kafka.Name))
		fmt.Fprintf(out, "\t\tTopic: %s\n", fastly.ToValue(kafka.Topic))
		fmt.Fprintf(out, "\t\tBrokers: %s\n", fastly.ToValue(kafka.Brokers))
		fmt.Fprintf(out, "\t\tRequired acks: %s\n", fastly.ToValue(kafka.RequiredACKs))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(kafka.CompressionCodec))
		fmt.Fprintf(out, "\t\tUse TLS: %t\n", fastly.ToValue(kafka.UseTLS))
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", fastly.ToValue(kafka.TLSCACert))
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", fastly.ToValue(kafka.TLSClientCert))
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", fastly.ToValue(kafka.TLSClientKey))
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", fastly.ToValue(kafka.TLSHostname))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(kafka.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(kafka.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(kafka.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(kafka.Placement))
		fmt.Fprintf(out, "\t\tParse log key-values: %t\n", fastly.ToValue(kafka.ParseLogKeyvals))
		fmt.Fprintf(out, "\t\tMax batch size: %d\n", fastly.ToValue(kafka.RequestMaxBytes))
		fmt.Fprintf(out, "\t\tSASL authentication method: %s\n", fastly.ToValue(kafka.AuthMethod))
		fmt.Fprintf(out, "\t\tSASL authentication username: %s\n", fastly.ToValue(kafka.User))
		fmt.Fprintf(out, "\t\tSASL authentication password: %s\n", fastly.ToValue(kafka.Password))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(kafka.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
