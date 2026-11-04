package kafka

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a Kafka logging endpoint.
type DescribeCommand struct {
	argparser.Base

	Input          fastly.GetKafkaInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"The name of the Kafka logging object"`
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

	o, err := c.Globals.APIClient.GetKafka(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Brokers":                      fastly.ToValue(o.Brokers),
		"Compression codec":            fastly.ToValue(o.CompressionCodec),
		"Format version":               fastly.ToValue(o.FormatVersion),
		"Format":                       fastly.ToValue(o.Format),
		"Max batch size":               fastly.ToValue(o.RequestMaxBytes),
		"Name":                         fastly.ToValue(o.Name),
		"Parse log key-values":         fastly.ToValue(o.ParseLogKeyvals),
		"Placement":                    fastly.ToValue(o.Placement),
		"Processing region":            fastly.ToValue(o.ProcessingRegion),
		"Required acks":                fastly.ToValue(o.RequiredACKs),
		"Response condition":           fastly.ToValue(o.ResponseCondition),
		"SASL authentication method":   fastly.ToValue(o.AuthMethod),
		"SASL authentication password": fastly.ToValue(o.Password),
		"SASL authentication username": fastly.ToValue(o.User),
		"TLS CA certificate":           fastly.ToValue(o.TLSCACert),
		"TLS client certificate":       fastly.ToValue(o.TLSClientCert),
		"TLS client key":               fastly.ToValue(o.TLSClientKey),
		"TLS hostname":                 fastly.ToValue(o.TLSHostname),
		"Topic":                        fastly.ToValue(o.Topic),
		"Use TLS":                      fastly.ToValue(o.UseTLS),
		"Version":                      fastly.ToValue(o.ServiceVersion),
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(o.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
