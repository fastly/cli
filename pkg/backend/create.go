package backend

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create backends.
type CreateCommand struct {
	common.Base
	Input fastly.CreateBackendInput

	// We must store all of the boolean flags seperatly to the input structure
	// so they can be casted to go-fastly's custom `Compatibool` type later.
	AutoLoadbalance bool
	UseSSL          bool
	SSLCheckCert    bool
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("create", "Create a backend on a Fastly service version").Alias("add")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.Service)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "Backend name").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("address", "A hostname, IPv4, or IPv6 address for the backend").Required().StringVar(&c.Input.Address)

	c.CmdClause.Flag("comment", "A descriptive note").StringVar(&c.Input.Comment)
	c.CmdClause.Flag("port", "Port number of the address").UintVar(&c.Input.Port)
	c.CmdClause.Flag("override-host", "The hostname to override the Host header").StringVar(&c.Input.OverrideHost)
	c.CmdClause.Flag("connect-timeout", "How long to wait for a timeout in milliseconds").UintVar(&c.Input.ConnectTimeout)
	c.CmdClause.Flag("max-conn", "Maximum number of connections").UintVar(&c.Input.MaxConn)
	c.CmdClause.Flag("first-byte-timeout", "How long to wait for the first bytes in milliseconds").UintVar(&c.Input.FirstByteTimeout)
	c.CmdClause.Flag("between-bytes-timeout", "How long to wait between bytes in milliseconds").UintVar(&c.Input.BetweenBytesTimeout)
	c.CmdClause.Flag("auto-loadbalance", "Whether or not this backend should be automatically load balanced").BoolVar(&c.AutoLoadbalance)
	c.CmdClause.Flag("weight", "Weight used to load balance this backend against others").UintVar(&c.Input.Weight)
	c.CmdClause.Flag("request-condition", "Condition, which if met, will select this backend during a request").StringVar(&c.Input.RequestCondition)
	c.CmdClause.Flag("healthcheck", "The name of the healthcheck to use with this backend").StringVar(&c.Input.HealthCheck)
	c.CmdClause.Flag("shield", "The shield POP designated to reduce inbound load on this origin by serving the cached data to the rest of the network").StringVar(&c.Input.Shield)
	c.CmdClause.Flag("use-ssl", "Whether or not to use SSL to reach the backend").BoolVar(&c.UseSSL)
	c.CmdClause.Flag("ssl-check-cert", "Be strict on checking SSL certs").BoolVar(&c.SSLCheckCert)
	c.CmdClause.Flag("ssl-ca-cert", "CA certificate attached to origin").StringVar(&c.Input.SSLCACert)
	c.CmdClause.Flag("ssl-client-cert", "Client certificate attached to origin").StringVar(&c.Input.SSLClientCert)
	c.CmdClause.Flag("ssl-client-key", "Client key attached to origin").StringVar(&c.Input.SSLClientKey)
	c.CmdClause.Flag("ssl-cert-hostname", "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all.").StringVar(&c.Input.SSLCertHostname)
	c.CmdClause.Flag("ssl-sni-hostname", "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all.").StringVar(&c.Input.SSLSNIHostname)
	c.CmdClause.Flag("min-tls-version", "Minimum allowed TLS version on SSL connections to this backend").StringVar(&c.Input.MinTLSVersion)
	c.CmdClause.Flag("max-tls-version", "Maximum allowed TLS version on SSL connections to this backend").StringVar(&c.Input.MaxTLSVersion)
	c.CmdClause.Flag("ssl-ciphers", "List of OpenSSL ciphers (see https://www.openssl.org/docs/man1.0.2/man1/ciphers for details)").StringsVar(&c.Input.SSLCiphers)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	// Sadly, go-fastly uses custom a `Compatibool` type as a boolean value that
	// marshalls to 0/1 instead of true/false for compatability with the API.
	// Therefore, we need to cast our real flag bool to a fastly.Compatibool.
	c.Input.AutoLoadbalance = fastly.CBool(c.AutoLoadbalance)
	c.Input.UseSSL = fastly.CBool(c.UseSSL)
	c.Input.SSLCheckCert = fastly.CBool(c.SSLCheckCert)

	b, err := c.Globals.Client.CreateBackend(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created backend %s (service %s version %d)", b.Name, b.ServiceID, b.Version)
	return nil
}
