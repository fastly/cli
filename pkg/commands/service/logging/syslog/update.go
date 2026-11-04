package syslog

import (
	"context"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Syslog logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	EndpointName   string                           `name:"name" short:"n" required:"" help:"The name of the Syslog logging object"`
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	Address           argparser.OptionalString    `help:"A hostname or IPv4 address"`
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	MessageType       argparser.OptionalString    `help:"How the message should be formatted. One of: classic (default), loggly, logplex or blank"`
	NewName           argparser.OptionalString    `help:"New name of the Syslog logging object"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	Port              argparser.OptionalInt       `help:"The port number"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to syslog. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	TLSCACert         argparser.OptionalString    `name:"tls-ca-cert" help:"A secure certificate to authenticate the server with. Must be in PEM format"`
	TLSClientCert     argparser.OptionalString    `help:"The client certificate used to make authenticated requests. Must be in PEM format"`
	TLSClientKey      argparser.OptionalString    `help:"The client private key used to make authenticated requests. Must be in PEM format"`
	TLSHostname       argparser.OptionalString    `help:"Used during the TLS handshake to validate the certificate"`
	Token             argparser.OptionalString    `name:"auth-token" help:"Whether to prepend each message with a specific token"`
	UseTLS            argparser.OptionalBool      `help:"Whether to use TLS for secure logging. Can be either true or false"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.ServiceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateSyslogInput, error) {
	input := fastly.UpdateSyslogInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.Address.WasSet {
		input.Address = &c.Address.Value
	}

	if c.Port.WasSet {
		input.Port = &c.Port.Value
	}

	if c.UseTLS.WasSet {
		input.UseTLS = new(fastly.Compatibool(c.UseTLS.Value))
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = &c.TLSCACert.Value
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = &c.TLSHostname.Value
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = &c.TLSClientCert.Value
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = &c.TLSClientKey.Value
	}

	if c.Token.WasSet {
		input.Token = &c.Token.Value
	}

	if c.Format.WasSet {
		input.Format = new(argparser.Content(c.Format.Value))
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.Placement.WasSet {
		if strings.TrimSpace(c.Placement.Value) == "" {
			input.Placement = fastly.NullValue[string]()
		} else {
			input.Placement = fastly.NewNullable(c.Placement.Value)
		}
	}

	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	syslog, err := c.Globals.APIClient.UpdateSyslog(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Updated Syslog logging endpoint %s (service %s version %d)",
		fastly.ToValue(syslog.Name),
		fastly.ToValue(syslog.ServiceID),
		fastly.ToValue(syslog.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *UpdateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "address", "auth-token", "autoclone", "format",
		"format-version", "new-name", "message-type", "placement",
		"processing-region", "port", "service-id", "service-name",
		"response-condition", "tls-ca-cert", "tls-client-cert",
		"tls-client-key", "tls-hostname", "use-tls",
	}
}
