package backend

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update backends.
type UpdateCommand struct {
	common.Base
	Input fastly.GetBackendInput

	NewName             common.OptionalString
	Comment             common.OptionalString
	Address             common.OptionalString
	Port                common.OptionalUint
	OverrideHost        common.OptionalString
	ConnectTimeout      common.OptionalUint
	MaxConn             common.OptionalUint
	FirstByteTimeout    common.OptionalUint
	BetweenBytesTimeout common.OptionalUint
	AutoLoadbalance     common.OptionalBool
	Weight              common.OptionalUint
	RequestCondition    common.OptionalString
	HealthCheck         common.OptionalString
	Hostname            common.OptionalString
	Shield              common.OptionalString
	UseSSL              common.OptionalBool
	SSLCheckCert        common.OptionalBool
	SSLCACert           common.OptionalString
	SSLClientCert       common.OptionalString
	SSLClientKey        common.OptionalString
	SSLCertHostname     common.OptionalString
	SSLSNIHostname      common.OptionalString
	MinTLSVersion       common.OptionalString
	MaxTLSVersion       common.OptionalString
	SSLCiphers          common.OptionalStringSlice
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update a backend on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.Service)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "backend name").Short('n').Required().StringVar(&c.Input.Name)

	c.CmdClause.Flag("new-name", "New backend name").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	c.CmdClause.Flag("address", "A hostname, IPv4, or IPv6 address for the backend").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "Port number of the address").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("override-host", "The hostname to override the Host header").Action(c.OverrideHost.Set).StringVar(&c.OverrideHost.Value)
	c.CmdClause.Flag("connect-timeout", "How long to wait for a timeout in milliseconds").Action(c.ConnectTimeout.Set).UintVar(&c.ConnectTimeout.Value)
	c.CmdClause.Flag("max-conn", "Maximum number of connections").Action(c.MaxConn.Set).UintVar(&c.MaxConn.Value)
	c.CmdClause.Flag("first-byte-timeout", "How long to wait for the first bytes in milliseconds").Action(c.FirstByteTimeout.Set).UintVar(&c.MaxConn.Value)
	c.CmdClause.Flag("between-bytes-timeout", "How long to wait between bytes in milliseconds").Action(c.BetweenBytesTimeout.Set).UintVar(&c.BetweenBytesTimeout.Value)
	c.CmdClause.Flag("auto-loadbalance", "Whether or not this backend should be automatically load balanced").Action(c.AutoLoadbalance.Set).BoolVar(&c.AutoLoadbalance.Value)
	c.CmdClause.Flag("weight", "Weight used to load balance this backend against others").Action(c.Weight.Set).UintVar(&c.Weight.Value)
	c.CmdClause.Flag("request-condition", "condition, which if met, will select this backend during a request").Action(c.RequestCondition.Set).StringVar(&c.RequestCondition.Value)
	c.CmdClause.Flag("healthcheck", "The name of the healthcheck to use with this backend").Action(c.HealthCheck.Set).StringVar(&c.HealthCheck.Value)
	c.CmdClause.Flag("shield", "The shield POP designated to reduce inbound load on this origin by serving the cached data to the rest of the network").Action(c.Shield.Set).StringVar(&c.Shield.Value)
	c.CmdClause.Flag("use-ssl", "Whether or not to use SSL to reach the backend").Action(c.UseSSL.Set).BoolVar(&c.UseSSL.Value)
	c.CmdClause.Flag("ssl-check-cert", "Be strict on checking SSL certs").Action(c.SSLCheckCert.Set).BoolVar(&c.SSLCheckCert.Value)
	c.CmdClause.Flag("ssl-ca-cert", "CA certificate attached to origin").Action(c.SSLCACert.Set).StringVar(&c.SSLCACert.Value)
	c.CmdClause.Flag("ssl-client-cert", "Client certificate attached to origin").Action(c.SSLClientCert.Set).StringVar(&c.SSLClientCert.Value)
	c.CmdClause.Flag("ssl-client-key", "Client key attached to origin").Action(c.SSLClientKey.Set).StringVar(&c.SSLClientKey.Value)
	c.CmdClause.Flag("ssl-cert-hostname", "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all.").Action(c.SSLCertHostname.Set).StringVar(&c.SSLCertHostname.Value)
	c.CmdClause.Flag("ssl-sni-hostname", "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all.").Action(c.SSLSNIHostname.Set).StringVar(&c.SSLSNIHostname.Value)
	c.CmdClause.Flag("min-tls-version", "Minimum allowed TLS version on SSL connections to this backend").Action(c.MinTLSVersion.Set).StringVar(&c.MinTLSVersion.Value)
	c.CmdClause.Flag("max-tls-version", "Maximum allowed TLS version on SSL connections to this backend").Action(c.MaxTLSVersion.Set).StringVar(&c.MaxTLSVersion.Value)
	c.CmdClause.Flag("ssl-ciphers", "List of OpenSSL ciphers (see https://www.openssl.org/docs/man1.0.2/man1/ciphers for details)").Action(c.SSLCiphers.Set).StringsVar(&c.SSLCiphers.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	b, err := c.Globals.Client.GetBackend(&c.Input)
	if err != nil {
		return err
	}

	// Copy existing values from GET to UpdateBackendInput strcuture
	input := &fastly.UpdateBackendInput{
		Service:             b.ServiceID,
		Version:             b.Version,
		Name:                b.Name,
		NewName:             b.Name,
		Address:             b.Address,
		Port:                b.Port,
		OverrideHost:        b.OverrideHost,
		ConnectTimeout:      b.ConnectTimeout,
		MaxConn:             b.MaxConn,
		FirstByteTimeout:    b.FirstByteTimeout,
		BetweenBytesTimeout: b.BetweenBytesTimeout,
		AutoLoadbalance:     fastly.CBool(b.AutoLoadbalance),
		Weight:              b.Weight,
		RequestCondition:    b.RequestCondition,
		HealthCheck:         b.HealthCheck,
		Shield:              b.Shield,
		UseSSL:              fastly.CBool(b.UseSSL),
		SSLCheckCert:        fastly.CBool(b.SSLCheckCert),
		SSLCACert:           b.SSLCACert,
		SSLClientCert:       b.SSLClientCert,
		SSLClientKey:        b.SSLClientKey,
		SSLCertHostname:     b.SSLCertHostname,
		SSLSNIHostname:      b.SSLSNIHostname,
		MinTLSVersion:       b.MinTLSVersion,
		MaxTLSVersion:       b.MaxTLSVersion,
		SSLCiphers:          b.SSLCiphers,
	}

	// Set values to existing ones to prevent accidental overwrite if empty.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Comment.Valid {
		input.Comment = c.Comment.Value
	}

	if c.Address.Valid {
		input.Address = c.Address.Value
	}

	if c.Port.Valid {
		input.Port = c.Port.Value
	}

	if c.OverrideHost.Valid {
		input.OverrideHost = c.OverrideHost.Value
	}

	if c.ConnectTimeout.Valid {
		input.ConnectTimeout = c.ConnectTimeout.Value
	}

	if c.MaxConn.Valid {
		input.MaxConn = c.MaxConn.Value
	}

	if c.FirstByteTimeout.Valid {
		input.FirstByteTimeout = c.FirstByteTimeout.Value
	}

	if c.BetweenBytesTimeout.Valid {
		input.BetweenBytesTimeout = c.BetweenBytesTimeout.Value
	}

	if c.AutoLoadbalance.Valid {
		input.AutoLoadbalance = fastly.CBool(c.AutoLoadbalance.Value)
	}

	if c.Weight.Valid {
		input.Weight = c.Weight.Value
	}

	if c.RequestCondition.Valid {
		input.RequestCondition = c.RequestCondition.Value
	}

	if c.HealthCheck.Valid {
		input.HealthCheck = c.HealthCheck.Value
	}

	if c.Shield.Valid {
		input.Shield = c.Shield.Value
	}

	if c.UseSSL.Valid {
		input.UseSSL = fastly.CBool(c.UseSSL.Value)
	}

	if c.SSLCheckCert.Valid {
		input.SSLCheckCert = fastly.CBool(c.SSLCheckCert.Value)
	}

	if c.SSLCACert.Valid {
		input.SSLCACert = c.SSLCACert.Value
	}

	if c.SSLClientCert.Valid {
		input.SSLClientCert = c.SSLClientCert.Value
	}

	if c.SSLClientKey.Valid {
		input.SSLClientKey = c.SSLClientKey.Value
	}

	if c.SSLCertHostname.Valid {
		input.SSLCertHostname = c.SSLCertHostname.Value
	}

	if c.SSLSNIHostname.Valid {
		input.SSLSNIHostname = c.SSLSNIHostname.Value
	}

	if c.MinTLSVersion.Valid {
		input.MinTLSVersion = c.MinTLSVersion.Value
	}

	if c.MaxTLSVersion.Valid {
		input.MaxTLSVersion = c.MaxTLSVersion.Value
	}

	if c.SSLCiphers.Valid {
		input.SSLCiphers = c.SSLCiphers.Value
	}

	b, err = c.Globals.Client.UpdateBackend(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated backend %s (service %s version %d)", b.Name, b.ServiceID, b.Version)
	return nil
}
