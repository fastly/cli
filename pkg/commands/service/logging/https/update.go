package https

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

// UpdateCommand calls the Fastly API to update an HTTPS logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	EndpointName   string                           `name:"name" short:"n" required:"" help:"The name of the HTTPS logging object"` // Can't shadow argparser.Base method Name().
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	CompressionCodec  argparser.OptionalString    `kong:"-"`
	ContentType       argparser.OptionalString    `help:"Content type of the header sent with the request"`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	GzipLevel         argparser.OptionalInt       `kong:"-"`
	HeaderName        argparser.OptionalString    `help:"Name of the custom header sent with the request"`
	HeaderValue       argparser.OptionalString    `help:"Value of the custom header sent with the request"`
	JSONFormat        argparser.OptionalString    `help:"Enforces valid JSON formatting for log entries. Can be disabled 0, array of json (wraps JSON log batches in an array) 1, or newline delimited json (places each JSON log entry onto a new line in a batch) 2"`
	MessageType       argparser.OptionalString    `help:"How the message should be formatted. One of: classic (default), loggly, logplex or blank"`
	Method            argparser.OptionalString    `help:"HTTP method used for request. Can be POST or PUT. Defaults to POST if not specified"`
	NewName           argparser.OptionalString    `help:"New name of the HTTPS logging object"`
	Period            argparser.OptionalInt       `help:"How frequently log files are finalized so they can be available for reading (in seconds, default 3600)"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to HTTPS. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	RequestMaxBytes   argparser.OptionalInt       `help:"Maximum size of log batch, if non-zero. Defaults to 100MB"`
	RequestMaxEntries argparser.OptionalInt       `help:"Maximum number of logs to append to a batch, if non-zero. Defaults to 10k"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	TLSCACert         argparser.OptionalString    `name:"tls-ca-cert" help:"A secure certificate to authenticate the server with. Must be in PEM format"`
	TLSClientCert     argparser.OptionalString    `help:"The client certificate used to make authenticated requests. Must be in PEM format"`
	TLSClientKey      argparser.OptionalString    `help:"The client private key used to make authenticated requests. Must be in PEM format"`
	TLSHostname       argparser.OptionalString    `help:"Used during the TLS handshake to validate the certificate"`
	URL               argparser.OptionalString    `help:"URL that log data will be sent to. Must use the https protocol"`

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
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateHTTPSInput, error) {
	input := fastly.UpdateHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}

	if c.ContentType.WasSet {
		input.ContentType = &c.ContentType.Value
	}

	if c.JSONFormat.WasSet {
		input.JSONFormat = &c.JSONFormat.Value
	}

	if c.HeaderName.WasSet {
		input.HeaderName = &c.HeaderName.Value
	}

	if c.HeaderValue.WasSet {
		input.HeaderValue = &c.HeaderValue.Value
	}

	if c.Method.WasSet {
		input.Method = &c.Method.Value
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = &c.RequestMaxEntries.Value
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = &c.RequestMaxBytes.Value
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

	if c.Period.WasSet {
		input.Period = &c.Period.Value
	}

	if c.Placement.WasSet {
		if strings.TrimSpace(c.Placement.Value) == "" {
			input.Placement = fastly.NullValue[string]()
		} else {
			input.Placement = fastly.NewNullable(c.Placement.Value)
		}
	}

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
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

	https, err := c.Globals.APIClient.UpdateHTTPS(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Updated HTTPS logging endpoint %s (service %s version %d)",
		fastly.ToValue(https.Name),
		fastly.ToValue(https.ServiceID),
		fastly.ToValue(https.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *UpdateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "autoclone", "content-type", "format",
		"format-version", "header-name", "header-value", "json-format",
		"message-type", "method", "new-name", "period", "placement",
		"processing-region", "request-max-bytes", "request-max-entries",
		"response-condition", "service-id", "service-name", "tls-ca-cert",
		"tls-client-cert", "tls-client-key", "tls-hostname", "url",
	}
}
