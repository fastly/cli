package kafka

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v6/fastly"
)

// DescribeCommand calls the Fastly API to describe a Kafka logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetKafkaInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a Kafka logging endpoint on a Fastly service version").Alias("get")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.CmdClause.Flag("name", "The name of the Kafka logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	kafka, err := c.Globals.APIClient.GetKafka(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(kafka)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", kafka.ServiceID)
	}
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
