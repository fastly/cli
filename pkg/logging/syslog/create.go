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

// CreateCommand calls the Fastly API to create Amazon Syslog logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateSyslogInput

	// We must store all of the boolean flags seperatly to the input structure
	// so they can be casted to go-fastly's custom `Compatibool` type later.
	useTLS bool
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an Syslog logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Syslog logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)

	c.CmdClause.Flag("address", "A hostname or IPv4 address").Required().StringVar(&c.Input.Address)

	c.CmdClause.Flag("port", "The port number").UintVar(&c.Input.Port)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").BoolVar(&c.useTLS)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").StringVar(&c.Input.TLSCACert)
	c.CmdClause.Flag("tls-hostname", "Used during the TLS handshake to validate the certificate").StringVar(&c.Input.TLSHostname)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").StringVar(&c.Input.TLSClientCert)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").StringVar(&c.Input.TLSClientKey)
	c.CmdClause.Flag("auth-token", "Whether to prepend each message with a specific token").StringVar(&c.Input.Token)
	c.CmdClause.Flag("format", "Apache style log formatting").StringVar(&c.Input.Format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").UintVar(&c.Input.FormatVersion)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").StringVar(&c.Input.MessageType)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").StringVar(&c.Input.ResponseCondition)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").StringVar(&c.Input.Placement)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	// Sadly, go-fastly uses custom a `Compatibool` type as a boolean value that
	// marshalls to 0/1 instead of true/false for compatability with the API.
	// Therefore, we need to cast our real flag bool to a fastly.Compatibool.
	c.Input.UseTLS = fastly.CBool(c.useTLS)

	d, err := c.Globals.Client.CreateSyslog(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Syslog logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
