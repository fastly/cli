package backend

import (
	"context"
	"io"
	"net"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create backends.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion
	Version        string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	Address             argparser.OptionalString        `help:"A hostname, IPv4, or IPv6 address for the backend"`
	AutoClone           argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	AutoLoadBalance     argparser.OptionalBool          `name:"auto-loadbalance" help:"Whether or not this backend should be automatically load balanced"`
	BetweenBytesTimeout argparser.OptionalInt           `help:"How long to wait between bytes in milliseconds"`
	Comment             argparser.OptionalString        `help:"A descriptive note"`
	ConnectTimeout      argparser.OptionalInt           `help:"How long to wait for a timeout in milliseconds"`
	FirstByteTimeout    argparser.OptionalInt           `help:"How long to wait for the first bytes in milliseconds"`
	HealthCheck         argparser.OptionalString        `name:"healthcheck" help:"The name of the healthcheck to use with this backend"`
	MaxConn             argparser.OptionalInt           `help:"Maximum number of connections"`
	MaxUse              argparser.OptionalInt           `help:"Maximum number of requests allowed over a single, pooled HTTP keepalive connection to this backend; 0 is treated as unlimited."`
	MaxLifetime         argparser.OptionalInt           `help:"Maximum time from creation (in milliseconds) that a pooled HTTP keepalive connection will be eligible for reuse; 0 is treated as unlimited."`
	MaxTLSVersion       argparser.OptionalString        `help:"Maximum allowed TLS version on SSL connections to this backend"`
	MinTLSVersion       argparser.OptionalString        `help:"Minimum allowed TLS version on SSL connections to this backend"`
	NameFlag            argparser.OptionalString        `name:"name" short:"n" help:"Backend name"`
	NoSSLCheckCert      argparser.OptionalBool          `help:"Skip checking SSL certs"`
	OverrideHost        argparser.OptionalString        `help:"The hostname to override the Host header"`
	Port                argparser.OptionalInt           `help:"Port number of the address"`
	PreferIPv6          argparser.OptionalString        `name:"prefer-ipv6" help:"Prefer IPv6 connections [true, false]"`
	RequestCondition    argparser.OptionalString        `help:"Condition, which if met, will select this backend during a request"`
	ServiceID           *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName         argparser.OptionalServiceNameID `help:"The name of the service"`
	Shield              argparser.OptionalString        `help:"The shield POP designated to reduce inbound load on this origin by serving the cached data to the rest of the network"`
	SSLCACert           argparser.OptionalString        `name:"ssl-ca-cert" help:"CA certificate attached to origin"`
	SSLCertHostname     argparser.OptionalString        `help:"Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all."`
	SSLCheckCert        argparser.OptionalBool          `help:"Be strict on checking SSL certs"`
	SSLCiphers          argparser.OptionalString        `help:"List of OpenSSL ciphers (https://www.openssl.org/docs/man1.0.2/man1/ciphers)"`
	SSLClientCert       argparser.OptionalString        `help:"Client certificate attached to origin"`
	SSLClientKey        argparser.OptionalString        `help:"Client key attached to origin"`
	SSLSNIHostname      argparser.OptionalString        `name:"ssl-sni-hostname" help:"Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all."`
	TCPKaEnable         argparser.OptionalString        `name:"tcp-ka-enabled" help:"Enable TCP keepalive probes [true, false]"`
	TCPKaInterval       argparser.OptionalInt           `help:"Configure how long to wait between sending each TCP keepalive probe."`
	TCPKaProbes         argparser.OptionalInt           `help:"Configure how many unacknowledged TCP keepalive probes to send before considering the connection dead."`
	TCPKaTime           argparser.OptionalInt           `help:"Configure how long to wait after the last sent data before sending TCP keepalive probes."`
	HTTPKaTime          argparser.OptionalInt           `help:"Configure how long to keep idle HTTP keepalive connections in the connection pool."`
	UseSSL              argparser.OptionalBool          `help:"Whether or not to use SSL to reach the backend"`
	Weight              argparser.OptionalInt           `help:"Weight used to load balance this backend against others"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
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
	input := fastly.CreateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
	}

	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Address.WasSet {
		input.Address = &c.Address.Value
	}
	if c.AutoLoadBalance.WasSet {
		input.AutoLoadbalance = new(fastly.Compatibool(c.AutoLoadBalance.Value))
	}
	if c.BetweenBytesTimeout.WasSet {
		input.BetweenBytesTimeout = &c.BetweenBytesTimeout.Value
	}
	if c.Comment.WasSet {
		input.Comment = &c.Comment.Value
	}
	if c.ConnectTimeout.WasSet {
		input.ConnectTimeout = &c.ConnectTimeout.Value
	}
	if c.FirstByteTimeout.WasSet {
		input.FirstByteTimeout = &c.FirstByteTimeout.Value
	}
	if c.HealthCheck.WasSet {
		input.HealthCheck = &c.HealthCheck.Value
	}
	if c.MaxConn.WasSet {
		input.MaxConn = &c.MaxConn.Value
	}
	if c.MaxUse.WasSet {
		input.MaxUse = &c.MaxUse.Value
	}
	if c.MaxLifetime.WasSet {
		input.MaxLifetime = &c.MaxLifetime.Value
	}
	if c.MaxTLSVersion.WasSet {
		input.MaxTLSVersion = &c.MaxTLSVersion.Value
	}
	if c.MinTLSVersion.WasSet {
		input.MinTLSVersion = &c.MinTLSVersion.Value
	}
	if c.NoSSLCheckCert.WasSet {
		input.SSLCheckCert = new(fastly.Compatibool(false))
	}
	if c.OverrideHost.WasSet {
		input.OverrideHost = &c.OverrideHost.Value
	}
	if c.PreferIPv6.WasSet {
		preferIPv6, err := argparser.ConvertBoolFromStringFlag(c.PreferIPv6.Value, "prefer-ipv6")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		input.PreferIPv6 = new(fastly.Compatibool(*preferIPv6))
	}
	if c.RequestCondition.WasSet {
		input.RequestCondition = &c.RequestCondition.Value
	}
	if c.Shield.WasSet {
		input.Shield = &c.Shield.Value
	}
	if c.SSLCACert.WasSet {
		input.SSLCACert = &c.SSLCACert.Value
	}
	if c.SSLCertHostname.WasSet {
		input.SSLCertHostname = &c.SSLCertHostname.Value
	}
	if c.SSLCheckCert.WasSet {
		text.Deprecated("The Fastly API defaults `ssl_check_cert` to true. Use `--no-ssl-check-cert` to disable this setting.\n\n")
		input.SSLCheckCert = new(fastly.Compatibool(c.SSLCheckCert.Value))
	}
	if c.SSLCiphers.WasSet {
		input.SSLCiphers = &c.SSLCiphers.Value
	}
	if c.SSLClientCert.WasSet {
		input.SSLClientCert = &c.SSLClientCert.Value
	}
	if c.SSLClientKey.WasSet {
		input.SSLClientKey = &c.SSLClientKey.Value
	}
	if c.SSLSNIHostname.WasSet {
		input.SSLSNIHostname = &c.SSLSNIHostname.Value
	}
	if c.TCPKaEnable.WasSet {
		tcpKaEnable, err := argparser.ConvertBoolFromStringFlag(c.TCPKaEnable.Value, "tcp-ka-enabled")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		input.TCPKeepAliveEnable = tcpKaEnable
	}
	if c.TCPKaInterval.WasSet {
		input.TCPKeepAliveIntvl = &c.TCPKaInterval.Value
	}
	if c.TCPKaProbes.WasSet {
		input.TCPKeepAliveProbes = &c.TCPKaProbes.Value
	}
	if c.TCPKaTime.WasSet {
		input.TCPKeepAliveTime = &c.TCPKaTime.Value
	}
	if c.HTTPKaTime.WasSet {
		input.KeepAliveTime = &c.HTTPKaTime.Value
	}
	if c.Weight.WasSet {
		input.Weight = &c.Weight.Value
	}

	switch {
	case c.Port.WasSet:
		input.Port = &c.Port.Value
	case c.UseSSL.WasSet && c.UseSSL.Value:
		if c.Globals.Flags.Verbose {
			text.Warning(out, "Use-ssl was set but no port was specified, using default port 443\n\n")
		}
		input.Port = new(443)
	}

	if input.Address != nil && !c.OverrideHost.WasSet && !c.SSLCertHostname.WasSet && !c.SSLSNIHostname.WasSet {
		overrideHost, sslSNIHostname, sslCertHostname := SetBackendHostDefaults(*input.Address)
		if overrideHost != "" {
			input.OverrideHost = &overrideHost
		}
		input.SSLSNIHostname = &sslSNIHostname
		input.SSLCertHostname = &sslCertHostname
	} else {
		if c.OverrideHost.WasSet {
			input.OverrideHost = &c.OverrideHost.Value
		}
		if c.SSLCertHostname.WasSet {
			input.SSLCertHostname = &c.SSLCertHostname.Value
		}
		if c.SSLSNIHostname.WasSet {
			input.SSLSNIHostname = &c.SSLSNIHostname.Value
		}
	}

	b, err := c.Globals.APIClient.CreateBackend(context.TODO(), &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created backend %s (service %s version %d)", fastly.ToValue(b.Name), fastly.ToValue(b.ServiceID), fastly.ToValue(b.ServiceVersion))
	return nil
}

// SetBackendHostDefaults configures the OverrideHost and SSLSNIHostname fields.
//
// By default we set the override_host and ssl_sni_hostname properties of the
// Backend object to the hostname, unless the given input is an IP.
func SetBackendHostDefaults(address string) (overrideHost, sslSNIHostname, sslCertHostname string) {
	if _, err := net.LookupAddr(address); err != nil {
		overrideHost = address
	}
	if overrideHost != "" {
		sslSNIHostname = overrideHost
		sslCertHostname = overrideHost
	}
	return overrideHost, sslSNIHostname, sslCertHostname
}
