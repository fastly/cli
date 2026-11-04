package backend

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe a backend.
type DescribeCommand struct {
	argparser.Base

	Input          fastly.GetBackendInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Name of backend"`
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

	o, err := c.Globals.APIClient.GetBackend(context.TODO(), &c.Input)
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
	fmt.Fprintf(out, "Prefer IPv6: %v\n", fastly.ToValue(b.PreferIPv6))
	fmt.Fprintf(out, "Override host: %v\n", fastly.ToValue(b.OverrideHost))
	fmt.Fprintf(out, "Connect timeout: %v\n", fastly.ToValue(b.ConnectTimeout))
	fmt.Fprintf(out, "Max connections: %v\n", fastly.ToValue(b.MaxConn))
	fmt.Fprintf(out, "Max connection use: %v\n", fastly.ToValue(b.MaxUse))
	fmt.Fprintf(out, "Max connection lifetime: %v\n", fastly.ToValue(b.MaxLifetime))
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
	fmt.Fprintf(out, "HTTP KeepAlive Timeout: %v\n", fastly.ToValue(b.KeepAliveTime))
	if b.TCPKeepAliveEnable == nil {
		fmt.Fprintf(out, "TCP KeepAlive Enabled: unset\n")
	} else {
		fmt.Fprintf(out, "TCP KeepAlive Enabled: %v\n", fastly.ToValue(b.TCPKeepAliveEnable))
	}
	fmt.Fprintf(out, "TCP KeepAlive Interval: %v\n", fastly.ToValue(b.TCPKeepAliveIntvl))
	fmt.Fprintf(out, "TCP KeepAlive Probes: %v\n", fastly.ToValue(b.TCPKeepAliveProbes))
	fmt.Fprintf(out, "TCP KeepAlive Timeout: %v\n", fastly.ToValue(b.TCPKeepAliveTime))

	return nil
}
