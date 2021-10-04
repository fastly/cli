package backend

import (
	_ "embed"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

//go:embed notes/update.txt
var updateNote string

// UpdateCommand calls the Fastly API to update backends.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetBackendInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	NewName             cmd.OptionalString
	Comment             cmd.OptionalString
	Address             cmd.OptionalString
	Port                cmd.OptionalUint
	OverrideHost        cmd.OptionalString
	ConnectTimeout      cmd.OptionalUint
	MaxConn             cmd.OptionalUint
	FirstByteTimeout    cmd.OptionalUint
	BetweenBytesTimeout cmd.OptionalUint
	AutoLoadbalance     cmd.OptionalBool
	Weight              cmd.OptionalUint
	RequestCondition    cmd.OptionalString
	HealthCheck         cmd.OptionalString
	Hostname            cmd.OptionalString
	Shield              cmd.OptionalString
	UseSSL              cmd.OptionalBool
	SSLCheckCert        cmd.OptionalBool
	SSLCACert           cmd.OptionalString
	SSLClientCert       cmd.OptionalString
	SSLClientKey        cmd.OptionalString
	SSLCertHostname     cmd.OptionalString
	SSLSNIHostname      cmd.OptionalString
	MinTLSVersion       cmd.OptionalString
	MaxTLSVersion       cmd.OptionalString
	SSLCiphers          cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update a backend on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
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
	c.CmdClause.Flag("ssl-ciphers", "Colon delimited list of OpenSSL ciphers (see https://www.openssl.org/docs/man1.0.2/man1/ciphers for details)").Action(c.SSLCiphers.Set).StringVar(&c.SSLCiphers.Value)
	return &c
}

// Notes displays useful contextual information.
func (c *UpdateCommand) Notes() string {
	return updateNote
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	b, err := c.Globals.Client.GetBackend(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	input := &fastly.UpdateBackendInput{
		ServiceID:      b.ServiceID,
		ServiceVersion: b.ServiceVersion,
		Name:           b.Name,
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Comment.WasSet {
		input.Comment = fastly.String(c.Comment.Value)
	}

	if c.Address.WasSet {
		input.Address = fastly.String(c.Address.Value)
	}

	if c.Port.WasSet {
		input.Port = fastly.Uint(c.Port.Value)
	}

	if c.OverrideHost.WasSet {
		input.OverrideHost = fastly.String(c.OverrideHost.Value)
	}

	if c.ConnectTimeout.WasSet {
		input.ConnectTimeout = fastly.Uint(c.ConnectTimeout.Value)
	}

	if c.MaxConn.WasSet {
		input.MaxConn = fastly.Uint(c.MaxConn.Value)
	}

	if c.FirstByteTimeout.WasSet {
		input.FirstByteTimeout = fastly.Uint(c.FirstByteTimeout.Value)
	}

	if c.BetweenBytesTimeout.WasSet {
		input.BetweenBytesTimeout = fastly.Uint(c.BetweenBytesTimeout.Value)
	}

	if c.AutoLoadbalance.WasSet {
		input.AutoLoadbalance = fastly.CBool(c.AutoLoadbalance.Value)
	}

	if c.Weight.WasSet {
		input.Weight = fastly.Uint(c.Weight.Value)
	}

	if c.RequestCondition.WasSet {
		input.RequestCondition = fastly.String(c.RequestCondition.Value)
	}

	if c.HealthCheck.WasSet {
		input.HealthCheck = fastly.String(c.HealthCheck.Value)
	}

	if c.Shield.WasSet {
		input.Shield = fastly.String(c.Shield.Value)
	}

	if c.UseSSL.WasSet {
		input.UseSSL = fastly.CBool(c.UseSSL.Value)
	}

	if c.SSLCheckCert.WasSet {
		input.SSLCheckCert = fastly.CBool(c.SSLCheckCert.Value)
	}

	if c.SSLCACert.WasSet {
		input.SSLCACert = fastly.String(c.SSLCACert.Value)
	}

	if c.SSLClientCert.WasSet {
		input.SSLClientCert = fastly.String(c.SSLClientCert.Value)
	}

	if c.SSLClientKey.WasSet {
		input.SSLClientKey = fastly.String(c.SSLClientKey.Value)
	}

	if c.SSLCertHostname.WasSet {
		input.SSLCertHostname = fastly.String(c.SSLCertHostname.Value)
	}

	if c.SSLSNIHostname.WasSet {
		input.SSLSNIHostname = fastly.String(c.SSLSNIHostname.Value)
	}

	if c.MinTLSVersion.WasSet {
		input.MinTLSVersion = fastly.String(c.MinTLSVersion.Value)
	}

	if c.MaxTLSVersion.WasSet {
		input.MaxTLSVersion = fastly.String(c.MaxTLSVersion.Value)
	}

	if c.SSLCiphers.WasSet {
		input.SSLCiphers = c.SSLCiphers.Value
	}

	b, err = c.Globals.Client.UpdateBackend(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Updated backend %s (service %s version %d)", b.Name, b.ServiceID, b.ServiceVersion)
	return nil
}
