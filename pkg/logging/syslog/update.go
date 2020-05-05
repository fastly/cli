package syslog

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Amazon Syslog logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	Input fastly.GetSyslogInput

	NewName           common.OptionalString
	Address           common.OptionalString
	Port              common.OptionalUint
	UseTLS            common.OptionalBool
	TLSCACert         common.OptionalString
	TLSHostname       common.OptionalString
	TLSClientCert     common.OptionalString
	TLSClientKey      common.OptionalString
	Token             common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	MessageType       common.OptionalString
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Syslog logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Syslog logging object").Short('n').Required().StringVar(&c.Input.Name)

	c.CmdClause.Flag("new-name", "New name of the Syslog logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("address", "A hostname or IPv4 address").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").Action(c.UseTLS.Set).BoolVar(&c.UseTLS.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-hostname", "Used during the TLS handshake to validate the certificate").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.TLSClientCert.Set).StringVar(&c.TLSClientCert.Value)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.TLSClientKey.Set).StringVar(&c.TLSClientKey.Value)
	c.CmdClause.Flag("auth-token", "Whether to prepend each message with a specific token").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	syslog, err := c.Globals.Client.GetSyslog(&c.Input)
	if err != nil {
		return err
	}

	input := &fastly.UpdateSyslogInput{
		Service:           syslog.ServiceID,
		Version:           syslog.Version,
		Name:              syslog.Name,
		NewName:           syslog.Name,
		Address:           syslog.Address,
		Port:              syslog.Port,
		UseTLS:            fastly.CBool(syslog.UseTLS),
		TLSCACert:         syslog.TLSCACert,
		TLSHostname:       syslog.TLSHostname,
		TLSClientCert:     syslog.TLSClientCert,
		TLSClientKey:      syslog.TLSClientKey,
		Token:             syslog.Token,
		Format:            syslog.Format,
		FormatVersion:     syslog.FormatVersion,
		MessageType:       syslog.MessageType,
		ResponseCondition: syslog.ResponseCondition,
		Placement:         syslog.Placement,
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Address.Valid {
		input.Address = c.Address.Value
	}

	if c.Port.Valid {
		input.Port = c.Port.Value
	}

	if c.UseTLS.Valid {
		input.UseTLS = fastly.CBool(c.UseTLS.Value)
	}

	if c.TLSCACert.Valid {
		input.TLSCACert = c.TLSCACert.Value
	}

	if c.TLSHostname.Valid {
		input.TLSHostname = c.TLSHostname.Value
	}

	if c.TLSClientCert.Valid {
		input.TLSClientCert = c.TLSClientCert.Value
	}

	if c.TLSClientKey.Valid {
		input.TLSClientKey = c.TLSClientKey.Value
	}

	if c.Token.Valid {
		input.Token = c.Token.Value
	}

	if c.Format.Valid {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.MessageType.Valid {
		input.MessageType = c.MessageType.Value
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.Placement.Valid {
		input.Placement = c.Placement.Value
	}

	syslog, err = c.Globals.Client.UpdateSyslog(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Syslog logging endpoint %s (service %s version %d)", syslog.Name, syslog.ServiceID, syslog.Version)
	return nil
}
