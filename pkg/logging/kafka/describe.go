package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe a Kafka logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetKafkaInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Kafka logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Kafka logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	kafka, err := c.Globals.Client.GetKafka(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", kafka.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", kafka.Version)
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

	return nil
}
