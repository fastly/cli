package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Kafka logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListKafkasInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Kafka endpoints on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
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

	kafkas, err := c.Globals.Client.ListKafkas(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, kafka := range kafkas {
			tw.AddLine(kafka.ServiceID, kafka.ServiceVersion, kafka.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kafka := range kafkas {
		fmt.Fprintf(out, "\tKafka %d/%d\n", i+1, len(kafkas))
		fmt.Fprintf(out, "\t\tService ID: %s\n", kafka.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", kafka.ServiceVersion)
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
