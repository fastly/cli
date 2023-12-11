package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
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
	c.Input.ServiceVersion = serviceVersion.Number

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
			tw.AddLine(kafka.ServiceID, kafka.ServiceVersion, kafka.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kafka := range o {
		fmt.Fprintf(out, "\tKafka %d/%d\n", i+1, len(o))
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
