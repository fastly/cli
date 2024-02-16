package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// ListCommand calls the Fastly API to list Kafka logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListKafkasInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Kafka endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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

	o, err := c.Globals.APIClient.ListKafkas(&c.Input)
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
	}
	fmt.Fprintln(out)

	return nil
}
