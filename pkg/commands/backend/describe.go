package backend

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// DescribeCommand calls the Fastly API to describe a backend.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.GetBackendInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a backend on a Fastly service version").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "Name of backend").Short('n').Required().StringVar(&c.Input.Name)
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
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	o, err := c.Globals.APIClient.GetBackend(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, b *fastly.Backend) error {
	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(b.ServiceID))
	}
	fmt.Fprintf(out, "Service Version: %d\n\n", fastly.ToValue(b.ServiceVersion))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(b.Name))
	fmt.Fprintf(out, "Comment: %v\n", fastly.ToValue(b.Comment))
	fmt.Fprintf(out, "Address: %v\n", fastly.ToValue(b.Address))
	fmt.Fprintf(out, "Port: %v\n", fastly.ToValue(b.Port))
	fmt.Fprintf(out, "Override host: %v\n", fastly.ToValue(b.OverrideHost))
	fmt.Fprintf(out, "Connect timeout: %v\n", fastly.ToValue(b.ConnectTimeout))
	fmt.Fprintf(out, "Max connections: %v\n", fastly.ToValue(b.MaxConn))
	fmt.Fprintf(out, "First byte timeout: %v\n", fastly.ToValue(b.FirstByteTimeout))
	fmt.Fprintf(out, "Between bytes timeout: %v\n", fastly.ToValue(b.BetweenBytesTimeout))
	fmt.Fprintf(out, "Auto loadbalance: %v\n", fastly.ToValue(b.AutoLoadbalance))
	fmt.Fprintf(out, "Weight: %v\n", fastly.ToValue(b.Weight))
	fmt.Fprintf(out, "Healthcheck: %v\n", fastly.ToValue(b.HealthCheck))
	fmt.Fprintf(out, "Shield: %v\n", fastly.ToValue(b.Shield))
	fmt.Fprintf(out, "Use SSL: %v\n", fastly.ToValue(b.UseSSL))
	fmt.Fprintf(out, "SSL check cert: %v\n", fastly.ToValue(b.SSLCheckCert))
	fmt.Fprintf(out, "SSL CA cert: %v\n", fastly.ToValue(b.SSLCACert))
	fmt.Fprintf(out, "SSL client cert: %v\n", fastly.ToValue(b.SSLClientCert))
	fmt.Fprintf(out, "SSL client key: %v\n", fastly.ToValue(b.SSLClientKey))
	fmt.Fprintf(out, "SSL cert hostname: %v\n", fastly.ToValue(b.SSLCertHostname))
	fmt.Fprintf(out, "SSL SNI hostname: %v\n", fastly.ToValue(b.SSLSNIHostname))
	fmt.Fprintf(out, "Min TLS version: %v\n", fastly.ToValue(b.MinTLSVersion))
	fmt.Fprintf(out, "Max TLS version: %v\n", fastly.ToValue(b.MaxTLSVersion))
	fmt.Fprintf(out, "SSL ciphers: %v\n", fastly.ToValue(b.SSLCiphers))

	return nil
}
