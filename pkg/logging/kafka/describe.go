package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Kafka logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetKafkaInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Kafka logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Kafka logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	kafka, err := c.Globals.Client.GetKafka(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", kafka.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", kafka.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", kafka.Name)
	fmt.Fprintf(out, "Topic: %s\n", kafka.Topic)
	fmt.Fprintf(out, "Brokers: %s\n", kafka.Brokers)
	fmt.Fprintf(out, "Required acks: %s\n", kafka.RequiredACKs)
	fmt.Fprintf(out, "Compression codec: %s\n", kafka.CompressionCodec)
	fmt.Fprintf(out, "Use TLS: %t\n", kafka.UseTLS)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", kafka.TLSCACert)
	fmt.Fprintf(out, "TLS client certificate: %s\n", kafka.TLSClientCert)
	fmt.Fprintf(out, "TLS client key: %s\n", kafka.TLSClientKey)
	fmt.Fprintf(out, "TLS hostname: %s\n", kafka.TLSHostname)
	fmt.Fprintf(out, "Format: %s\n", kafka.Format)
	fmt.Fprintf(out, "Format version: %d\n", kafka.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", kafka.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", kafka.Placement)
	fmt.Fprintf(out, "Parse log key-values: %t\n", kafka.ParseLogKeyvals)
	fmt.Fprintf(out, "Max batch size: %d\n", kafka.RequestMaxBytes)
	fmt.Fprintf(out, "SASL authentication method: %s\n", kafka.AuthMethod)
	fmt.Fprintf(out, "SASL authentication username: %s\n", kafka.User)
	fmt.Fprintf(out, "SASL authentication password: %s\n", kafka.Password)

	return nil
}
