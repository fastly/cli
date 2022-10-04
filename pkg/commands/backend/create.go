package backend

import (
	"io"
	"net"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand calls the Fastly API to create backends.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data

	input               fastly.CreateBackendInput
	port                cmd.OptionalUint
	connectTimeout      cmd.OptionalUint
	maxConn             cmd.OptionalUint
	firstByteTimeout    cmd.OptionalUint
	betweenBytesTimeout cmd.OptionalUint
	weight              cmd.OptionalUint

	// We must store all of the boolean flags separately to the input structure
	// so they can be casted to go-fastly's custom `Compatibool` type later.
	autoLoadBalance bool
	sslCheckCert    bool
	useSSL          bool

	autoClone       cmd.OptionalAutoClone
	overrideHost    cmd.OptionalString
	serviceName     cmd.OptionalServiceNameID
	serviceVersion  cmd.OptionalServiceVersion
	sslCertHostname cmd.OptionalString
	sslSNIHostname  cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("create", "Create a backend on a Fastly service version").Alias("add")
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
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Backend name").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("address", "A hostname, IPv4, or IPv6 address for the backend").Required().StringVar(&c.input.Address)
	c.CmdClause.Flag("comment", "A descriptive note").StringVar(&c.input.Comment)
	c.CmdClause.Flag("port", "Port number of the address").Action(c.port.Set).UintVar(&c.port.Value)
	c.CmdClause.Flag("override-host", "The hostname to override the Host header").Action(c.overrideHost.Set).StringVar(&c.overrideHost.Value)
	c.CmdClause.Flag("connect-timeout", "How long to wait for a timeout in milliseconds").Action(c.connectTimeout.Set).UintVar(&c.connectTimeout.Value)
	c.CmdClause.Flag("max-conn", "Maximum number of connections").Action(c.maxConn.Set).UintVar(&c.maxConn.Value)
	c.CmdClause.Flag("first-byte-timeout", "How long to wait for the first bytes in milliseconds").Action(c.firstByteTimeout.Set).UintVar(&c.firstByteTimeout.Value)
	c.CmdClause.Flag("between-bytes-timeout", "How long to wait between bytes in milliseconds").Action(c.betweenBytesTimeout.Set).UintVar(&c.betweenBytesTimeout.Value)
	c.CmdClause.Flag("auto-loadbalance", "Whether or not this backend should be automatically load balanced").BoolVar(&c.autoLoadBalance)
	c.CmdClause.Flag("weight", "Weight used to load balance this backend against others").Action(c.weight.Set).UintVar(&c.weight.Value)
	c.CmdClause.Flag("request-condition", "Condition, which if met, will select this backend during a request").StringVar(&c.input.RequestCondition)
	c.CmdClause.Flag("healthcheck", "The name of the healthcheck to use with this backend").StringVar(&c.input.HealthCheck)
	c.CmdClause.Flag("shield", "The shield POP designated to reduce inbound load on this origin by serving the cached data to the rest of the network").StringVar(&c.input.Shield)
	c.CmdClause.Flag("use-ssl", "Whether or not to use SSL to reach the backend").BoolVar(&c.useSSL)
	c.CmdClause.Flag("ssl-check-cert", "Be strict on checking SSL certs").BoolVar(&c.sslCheckCert)
	c.CmdClause.Flag("ssl-ca-cert", "CA certificate attached to origin").StringVar(&c.input.SSLCACert)
	c.CmdClause.Flag("ssl-client-cert", "Client certificate attached to origin").StringVar(&c.input.SSLClientCert)
	c.CmdClause.Flag("ssl-client-key", "Client key attached to origin").StringVar(&c.input.SSLClientKey)
	c.CmdClause.Flag("ssl-cert-hostname", "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all.").Action(c.sslCertHostname.Set).StringVar(&c.sslCertHostname.Value)
	c.CmdClause.Flag("ssl-sni-hostname", "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all.").Action(c.sslSNIHostname.Set).StringVar(&c.sslSNIHostname.Value)
	c.CmdClause.Flag("min-tls-version", "Minimum allowed TLS version on SSL connections to this backend").StringVar(&c.input.MinTLSVersion)
	c.CmdClause.Flag("max-tls-version", "Maximum allowed TLS version on SSL connections to this backend").StringVar(&c.input.MaxTLSVersion)
	c.CmdClause.Flag("ssl-ciphers", "List of OpenSSL ciphers (https://www.openssl.org/docs/man1.0.2/man1/ciphers)").StringVar(&c.input.SSLCiphers)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	// Sadly, go-fastly uses custom a `Compatibool` type as a boolean value that
	// marshalls to 0/1 instead of true/false for compatibility with the API.
	// Therefore, we need to cast our real flag bool to a fastly.Compatibool.
	c.input.AutoLoadbalance = fastly.Compatibool(c.autoLoadBalance)
	c.input.UseSSL = fastly.Compatibool(c.useSSL)
	c.input.SSLCheckCert = fastly.Compatibool(c.sslCheckCert)

	switch {
	case c.port.WasSet:
		c.input.Port = fastly.Uint(c.port.Value)
	case c.useSSL:
		if c.Globals.Flag.Verbose {
			text.Warning(out, "Use-ssl was set but no port was specified, using default port 443")
		}
		c.input.Port = fastly.Uint(443)
	}

	if c.connectTimeout.WasSet {
		c.input.ConnectTimeout = fastly.Uint(c.connectTimeout.Value)
	}
	if c.maxConn.WasSet {
		c.input.MaxConn = fastly.Uint(c.maxConn.Value)
	}
	if c.firstByteTimeout.WasSet {
		c.input.FirstByteTimeout = fastly.Uint(c.firstByteTimeout.Value)
	}
	if c.betweenBytesTimeout.WasSet {
		c.input.BetweenBytesTimeout = fastly.Uint(c.betweenBytesTimeout.Value)
	}
	if c.weight.WasSet {
		c.input.Weight = fastly.Uint(c.weight.Value)
	}

	if !c.overrideHost.WasSet && !c.sslCertHostname.WasSet && !c.sslSNIHostname.WasSet {
		overrideHost, sslSNIHostname, sslCertHostname := SetBackendHostDefaults(c.input.Address)
		c.input.OverrideHost = overrideHost
		c.input.SSLSNIHostname = sslSNIHostname
		c.input.SSLCertHostname = sslCertHostname
	} else {
		if c.overrideHost.WasSet {
			c.input.OverrideHost = c.overrideHost.Value
		}
		if c.sslCertHostname.WasSet {
			c.input.SSLCertHostname = c.sslCertHostname.Value
		}
		if c.sslSNIHostname.WasSet {
			c.input.SSLSNIHostname = c.sslSNIHostname.Value
		}
	}

	b, err := c.Globals.APIClient.CreateBackend(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created backend %s (service %s version %d)", b.Name, b.ServiceID, b.ServiceVersion)
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
