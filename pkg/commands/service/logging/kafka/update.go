package kafka

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Kafka logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	EndpointName   string                           `name:"name" short:"n" required:"" help:"The name of the Kafka logging object"` // Can't shadow argparser.Base method Name().
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	AuthMethod        argparser.OptionalString    `hints:"plain,scram-sha-256,scram-sha-512" help:"SASL authentication method. Valid values are: plain, scram-sha-256, scram-sha-512"`
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Brokers           argparser.OptionalString    `help:"A comma-separated list of IP addresses or hostnames of Kafka brokers"`
	CompressionCodec  argparser.OptionalString    `help:"The codec used for compression of your logs. One of: gzip, snappy, lz4"`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	Index             argparser.OptionalString    `kong:"-"`
	NewName           argparser.OptionalString    `help:"New name of the Kafka logging object"`
	ParseLogKeyvals   argparser.OptionalBool      `negatable:"" help:"Parse key-value pairs within the log format"`
	Password          argparser.OptionalString    `help:"SASL authentication password. Required if --auth-method is specified"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to Kafka. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	RequestMaxBytes   argparser.OptionalInt       `name:"max-batch-size" help:"The maximum size of the log batch in bytes"`
	RequiredACKs      argparser.OptionalString    `name:"required-acks" help:"The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0\tNo servers need to respond. -1\tWait for all in-sync replicas to respond"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	TLSCACert         argparser.OptionalString    `name:"tls-ca-cert" help:"A secure certificate to authenticate the server with. Must be in PEM format"`
	TLSClientCert     argparser.OptionalString    `help:"The client certificate used to make authenticated requests. Must be in PEM format"`
	TLSClientKey      argparser.OptionalString    `help:"The client private key used to make authenticated requests. Must be in PEM format"`
	TLSHostname       argparser.OptionalString    `help:"Used during the TLS handshake to validate the certificate"`
	Topic             argparser.OptionalString    `help:"The Kafka topic to send logs to"`
	UseSASL           argparser.OptionalBool      `help:"Enable SASL authentication. Requires --auth-method, --username, and --password to be specified"`
	UseTLS            argparser.OptionalBool      `help:"Whether to use TLS for secure logging. Can be either true or false"`
	User              argparser.OptionalString    `name:"username" help:"SASL authentication username. Required if --auth-method is specified"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *UpdateCommand) AfterApply(ctx *kong.Context) error {
	if (c.AuthMethod.Value != "" || kongcompat.Given(ctx, "auth-method")) && !slices.Contains([]string{"plain", "scram-sha-256", "scram-sha-512"}, c.AuthMethod.Value) {
		return fmt.Errorf("enum value must be one of plain,scram-sha-256,scram-sha-512, got '%s'", c.AuthMethod.Value)
	}
	c.ServiceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateKafkaInput, error) {
	if c.UseSASL.WasSet && c.UseSASL.Value && (c.AuthMethod.Value == "" || c.User.Value == "" || c.Password.Value == "") {
		return nil, fmt.Errorf("the --auth-method, --username, and --password flags must be present when using the --use-sasl flag")
	}

	if !c.UseSASL.Value && (c.AuthMethod.Value != "" || c.User.Value != "" || c.Password.Value != "") {
		return nil, fmt.Errorf("the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified")
	}

	input := fastly.UpdateKafkaInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.Topic.WasSet {
		input.Topic = &c.Topic.Value
	}

	if c.Brokers.WasSet {
		input.Brokers = &c.Brokers.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}

	if c.RequiredACKs.WasSet {
		input.RequiredACKs = &c.RequiredACKs.Value
	}

	if c.UseTLS.WasSet {
		input.UseTLS = new(fastly.Compatibool(c.UseTLS.Value))
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = &c.TLSCACert.Value
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = &c.TLSClientCert.Value
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = &c.TLSClientKey.Value
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = &c.TLSHostname.Value
	}

	if c.Format.WasSet {
		input.Format = new(argparser.Content(c.Format.Value))
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
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

	if c.ParseLogKeyvals.WasSet {
		input.ParseLogKeyvals = new(fastly.Compatibool(c.ParseLogKeyvals.Value))
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = &c.RequestMaxBytes.Value
	}

	if c.UseSASL.WasSet && !c.UseSASL.Value {
		input.AuthMethod = new("")
		input.User = new("")
		input.Password = new("")
	}

	if c.AuthMethod.WasSet {
		input.AuthMethod = &c.AuthMethod.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.Password.WasSet {
		input.Password = &c.Password.Value
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

	kafka, err := c.Globals.APIClient.UpdateKafka(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Updated Kafka logging endpoint %s (service %s version %d)",
		fastly.ToValue(kafka.Name),
		fastly.ToValue(kafka.ServiceID),
		fastly.ToValue(kafka.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *UpdateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "autoclone", "auth-method", "brokers",
		"compression-codec", "format", "format-version", "max-batch-size",
		"new-name", "parse-log-keyvals", "password", "placement",
		"processing-region", "required-acks", "response-condition",
		"service-id", "service-name", "tls-ca-cert", "tls-client-cert",
		"tls-client-key", "tls-hostname", "topic", "use-sasl", "use-tls",
		"username",
	}
}
