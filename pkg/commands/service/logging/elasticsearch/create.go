package elasticsearch

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Elasticsearch logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	EndpointName      argparser.OptionalString    `name:"name" short:"n" help:"The name of the Elasticsearch logging object. Used as a primary key for API access"` // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	Index             argparser.OptionalString    `help:"The name of the Elasticsearch index to send documents (logs) to. The index must follow the Elasticsearch index format rules (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html). We support strftime (http://man7.org/linux/man-pages/man3/strftime.3.html) interpolated variables inside braces prefixed with a pound symbol. For example, #{%F} will interpolate as YYYY-MM-DD with today's date"`
	Password          argparser.OptionalString    `kong:"-"`
	Pipeline          argparser.OptionalString    `kong:"-"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to Elasticsearch. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	RequestMaxBytes   argparser.OptionalInt       `help:"Maximum size of log batch, if non-zero. Defaults to 100MB"`
	RequestMaxEntries argparser.OptionalInt       `help:"Maximum number of logs to append to a batch, if non-zero. Defaults to 10k"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	TLSCACert         argparser.OptionalString    `name:"tls-ca-cert" help:"A secure certificate to authenticate the server with. Must be in PEM format"`
	TLSClientCert     argparser.OptionalString    `help:"The client certificate used to make authenticated requests. Must be in PEM format"`
	TLSClientKey      argparser.OptionalString    `help:"The client private key used to make authenticated requests. Must be in PEM format"`
	TLSHostname       argparser.OptionalString    `help:"Used during the TLS handshake to validate the certificate"`
	URL               argparser.OptionalString    `help:"The URL to stream logs to. Must use HTTPS."`
	User              argparser.OptionalString    `kong:"-"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	PipelineFlag *string `name:"pipeline" help:"The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing. For example my_pipeline_id. Learn more about creating a pipeline in the Elasticsearch docs (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)"`
	ServiceID    *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.ServiceVersion.Value = c.Version
	if c.PipelineFlag != nil {
		c.Pipeline.Value = *c.PipelineFlag
		c.Password.WasSet = true
	}
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateElasticsearchInput, error) {
	var input fastly.CreateElasticsearchInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Index.WasSet {
		input.Index = &c.Index.Value
	}
	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}

	if c.Pipeline.WasSet {
		input.Pipeline = &c.Pipeline.Value
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = &c.RequestMaxEntries.Value
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = &c.RequestMaxBytes.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.Password.WasSet {
		input.Password = &c.Password.Value
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
		input.Placement = &c.Placement.Value
	}

	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	d, err := c.Globals.APIClient.CreateElasticsearch(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Created Elasticsearch logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *CreateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "autoclone", "format", "format-version", "index",
		"placement", "pipeline", "processing-region", "request-max-bytes",
		"request-max-entries", "response-condition", "service-id",
		"service-name", "tls-ca-cert", "tls-client-cert", "tls-client-key",
		"tls-hostname", "url",
	}
}
