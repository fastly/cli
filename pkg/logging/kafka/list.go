package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list Kafka logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListKafkasInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Kafka endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	kafkas, err := c.Globals.Client.ListKafkas(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, kafka := range kafkas {
			tw.AddLine(kafka.ServiceID, kafka.Version, kafka.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, kafka := range kafkas {
		fmt.Fprintf(out, "\tKafka %d/%d\n", i+1, len(kafkas))
		fmt.Fprintf(out, "\t\tService ID: %s\n", kafka.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", kafka.Version)
		fmt.Fprintf(out, "\t\tName: %s\n", kafka.Name)
		fmt.Fprintf(out, "\t\tTopic: %s\n", kafka.Topic)
		fmt.Fprintf(out, "\t\tBrokers: %s\n", kafka.Brokers)
		fmt.Fprintf(out, "\t\tRequired acks: %s\n", kafka.RequiredACKs)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", kafka.CompressionCodec)
		fmt.Fprintf(out, "\t\tUse TLS: %t\n", kafka.UseTLS)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", kafka.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", kafka.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", kafka.TLSClientKey)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", kafka.TLSHostname)
		fmt.Fprintf(out, "\t\tFormat: %s\n", kafka.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", kafka.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", kafka.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", kafka.Placement)
		fmt.Fprintf(out, "\t\tParse log key-values: %t\n", kafka.ParseLogKeyvals)
		fmt.Fprintf(out, "\t\tMax batch size: %d\n", kafka.RequestMaxBytes)
		fmt.Fprintf(out, "\t\tSASL authentication method: %s\n", kafka.AuthMethod)
		fmt.Fprintf(out, "\t\tSASL authentication username: %s\n", kafka.User)
		fmt.Fprintf(out, "\t\tSASL authentication password: %s\n", kafka.Password)
	}
	fmt.Fprintln(out)

	return nil
}
