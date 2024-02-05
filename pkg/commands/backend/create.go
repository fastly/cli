package backend

import (
	"io"
	"net"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create backends.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	address             argparser.OptionalString
	autoClone           argparser.OptionalAutoClone
	autoLoadBalance     argparser.OptionalBool
	betweenBytesTimeout argparser.OptionalInt
	comment             argparser.OptionalString
	connectTimeout      argparser.OptionalInt
	firstByteTimeout    argparser.OptionalInt
	healthCheck         argparser.OptionalString
	maxConn             argparser.OptionalInt
	maxTLSVersion       argparser.OptionalString
	minTLSVersion       argparser.OptionalString
	name                argparser.OptionalString
	noSSLCheckCert      argparser.OptionalBool
	overrideHost        argparser.OptionalString
	port                argparser.OptionalInt
	requestCondition    argparser.OptionalString
	serviceName         argparser.OptionalServiceNameID
	shield              argparser.OptionalString
	sslCACert           argparser.OptionalString
	sslCertHostname     argparser.OptionalString
	sslCheckCert        argparser.OptionalBool
	sslCiphers          argparser.OptionalString
	sslClientCert       argparser.OptionalString
	sslClientKey        argparser.OptionalString
	sslSNIHostname      argparser.OptionalString
	useSSL              argparser.OptionalBool
	weight              argparser.OptionalInt
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a backend on a Fastly service version").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.

	c.CmdClause.Flag("address", "A hostname, IPv4, or IPv6 address for the backend").Action(c.address.Set).StringVar(&c.address.Value)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("auto-loadbalance", "Whether or not this backend should be automatically load balanced").Action(c.autoLoadBalance.Set).BoolVar(&c.autoLoadBalance.Value)
	c.CmdClause.Flag("between-bytes-timeout", "How long to wait between bytes in milliseconds").Action(c.betweenBytesTimeout.Set).IntVar(&c.betweenBytesTimeout.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("connect-timeout", "How long to wait for a timeout in milliseconds").Action(c.connectTimeout.Set).IntVar(&c.connectTimeout.Value)
	c.CmdClause.Flag("first-byte-timeout", "How long to wait for the first bytes in milliseconds").Action(c.firstByteTimeout.Set).IntVar(&c.firstByteTimeout.Value)
	c.CmdClause.Flag("healthcheck", "The name of the healthcheck to use with this backend").Action(c.healthCheck.Set).StringVar(&c.healthCheck.Value)
	c.CmdClause.Flag("max-conn", "Maximum number of connections").Action(c.maxConn.Set).IntVar(&c.maxConn.Value)
	c.CmdClause.Flag("max-tls-version", "Maximum allowed TLS version on SSL connections to this backend").Action(c.maxTLSVersion.Set).StringVar(&c.maxTLSVersion.Value)
	c.CmdClause.Flag("min-tls-version", "Minimum allowed TLS version on SSL connections to this backend").Action(c.minTLSVersion.Set).StringVar(&c.minTLSVersion.Value)
	c.CmdClause.Flag("name", "Backend name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("no-ssl-check-cert", "Skip checking SSL certs").Action(c.noSSLCheckCert.Set).BoolVar(&c.noSSLCheckCert.Value)
	c.CmdClause.Flag("override-host", "The hostname to override the Host header").Action(c.overrideHost.Set).StringVar(&c.overrideHost.Value)
	c.CmdClause.Flag("port", "Port number of the address").Action(c.port.Set).IntVar(&c.port.Value)
	c.CmdClause.Flag("request-condition", "Condition, which if met, will select this backend during a request").Action(c.requestCondition.Set).StringVar(&c.requestCondition.Value)
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
	c.CmdClause.Flag("shield", "The shield POP designated to reduce inbound load on this origin by serving the cached data to the rest of the network").Action(c.shield.Set).StringVar(&c.shield.Value)
	c.CmdClause.Flag("ssl-ca-cert", "CA certificate attached to origin").Action(c.sslCACert.Set).StringVar(&c.sslCACert.Value)
	c.CmdClause.Flag("ssl-cert-hostname", "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all.").Action(c.sslCertHostname.Set).StringVar(&c.sslCertHostname.Value)
	c.CmdClause.Flag("ssl-check-cert", "Be strict on checking SSL certs").Action(c.sslCheckCert.Set).BoolVar(&c.sslCheckCert.Value)
	c.CmdClause.Flag("ssl-ciphers", "List of OpenSSL ciphers (https://www.openssl.org/docs/man1.0.2/man1/ciphers)").Action(c.sslCiphers.Set).StringVar(&c.sslCiphers.Value)
	c.CmdClause.Flag("ssl-client-cert", "Client certificate attached to origin").Action(c.sslClientCert.Set).StringVar(&c.sslClientCert.Value)
	c.CmdClause.Flag("ssl-client-key", "Client key attached to origin").Action(c.sslClientKey.Set).StringVar(&c.sslClientKey.Value)
	c.CmdClause.Flag("ssl-sni-hostname", "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all.").Action(c.sslSNIHostname.Set).StringVar(&c.sslSNIHostname.Value)
	c.CmdClause.Flag("use-ssl", "Whether or not to use SSL to reach the backend").Action(c.useSSL.Set).BoolVar(&c.useSSL.Value)
	c.CmdClause.Flag("weight", "Weight used to load balance this backend against others").Action(c.weight.Set).IntVar(&c.weight.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}
	input := fastly.CreateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
	}

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.address.WasSet {
		input.Address = &c.address.Value
	}
	if c.autoLoadBalance.WasSet {
		input.AutoLoadbalance = fastly.ToPointer(fastly.Compatibool(c.autoLoadBalance.Value))
	}
	if c.betweenBytesTimeout.WasSet {
		input.BetweenBytesTimeout = &c.betweenBytesTimeout.Value
	}
	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	if c.connectTimeout.WasSet {
		input.ConnectTimeout = &c.connectTimeout.Value
	}
	if c.firstByteTimeout.WasSet {
		input.FirstByteTimeout = &c.firstByteTimeout.Value
	}
	if c.healthCheck.WasSet {
		input.HealthCheck = &c.healthCheck.Value
	}
	if c.maxConn.WasSet {
		input.MaxConn = &c.maxConn.Value
	}
	if c.maxTLSVersion.WasSet {
		input.MaxTLSVersion = &c.maxTLSVersion.Value
	}
	if c.minTLSVersion.WasSet {
		input.MinTLSVersion = &c.minTLSVersion.Value
	}
	if c.noSSLCheckCert.WasSet {
		input.SSLCheckCert = fastly.ToPointer(fastly.Compatibool(false))
	}
	if c.overrideHost.WasSet {
		input.OverrideHost = &c.overrideHost.Value
	}
	if c.requestCondition.WasSet {
		input.RequestCondition = &c.requestCondition.Value
	}
	if c.shield.WasSet {
		input.Shield = &c.shield.Value
	}
	if c.sslCACert.WasSet {
		input.SSLCACert = &c.sslCACert.Value
	}
	if c.sslCertHostname.WasSet {
		input.SSLCertHostname = &c.sslCertHostname.Value
	}
	if c.sslCheckCert.WasSet {
		text.Deprecated(out, "The Fastly API defaults `ssl_check_cert` to true. Use `--no-ssl-check-cert` to disable this setting.\n\n")
		input.SSLCheckCert = fastly.ToPointer(fastly.Compatibool(c.sslCheckCert.Value))
	}
	if c.sslCiphers.WasSet {
		input.SSLCiphers = &c.sslCiphers.Value
	}
	if c.sslClientCert.WasSet {
		input.SSLClientCert = &c.sslClientCert.Value
	}
	if c.sslClientKey.WasSet {
		input.SSLClientKey = &c.sslClientKey.Value
	}
	if c.sslSNIHostname.WasSet {
		input.SSLSNIHostname = &c.sslSNIHostname.Value
	}
	if c.weight.WasSet {
		input.Weight = &c.weight.Value
	}

	switch {
	case c.port.WasSet:
		input.Port = &c.port.Value
	case c.useSSL.WasSet && c.useSSL.Value:
		if c.Globals.Flags.Verbose {
			text.Warning(out, "Use-ssl was set but no port was specified, using default port 443\n\n")
		}
		input.Port = fastly.ToPointer(443)
	}

	if input.Address != nil && !c.overrideHost.WasSet && !c.sslCertHostname.WasSet && !c.sslSNIHostname.WasSet {
		overrideHost, sslSNIHostname, sslCertHostname := SetBackendHostDefaults(*input.Address)
		if overrideHost != "" {
			input.OverrideHost = &overrideHost
		}
		input.SSLSNIHostname = &sslSNIHostname
		input.SSLCertHostname = &sslCertHostname
	} else {
		if c.overrideHost.WasSet {
			input.OverrideHost = &c.overrideHost.Value
		}
		if c.sslCertHostname.WasSet {
			input.SSLCertHostname = &c.sslCertHostname.Value
		}
		if c.sslSNIHostname.WasSet {
			input.SSLSNIHostname = &c.sslSNIHostname.Value
		}
	}

	b, err := c.Globals.APIClient.CreateBackend(&input)
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
