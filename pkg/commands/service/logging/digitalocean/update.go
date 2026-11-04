package digitalocean

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

// UpdateCommand calls the Fastly API to update a DigitalOcean Spaces logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	EndpointName   string                           `name:"name" short:"n" required:"" help:"The name of the DigitalOcean Spaces logging object"`
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	AccessKey         argparser.OptionalString    `help:"Your DigitalOcean Spaces account access key"`
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	BucketName        argparser.OptionalString    `name:"bucket" help:"The name of the DigitalOcean Space"`
	CompressionCodec  argparser.OptionalString    `help:"The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is \"gzip\", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error."`
	Domain            argparser.OptionalString    `help:"The domain of the DigitalOcean Spaces endpoint (default 'nyc3.digitaloceanspaces.com')"`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	GzipLevel         argparser.OptionalInt       `help:"What level of GZIP encoding to have when dumping logs (default 0, no compression)"`
	MessageType       argparser.OptionalString    `help:"How the message should be formatted. One of: classic (default), loggly, logplex or blank"`
	NewName           argparser.OptionalString    `help:"New name of the DigitalOcean Spaces logging object"`
	Path              argparser.OptionalString    `help:"The path to upload logs to"`
	Period            argparser.OptionalInt       `help:"How frequently log files are finalized so they can be available for reading (in seconds, default 3600)"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to DigitalOcean Spaces. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	PublicKey         argparser.OptionalString    `help:"A PGP public key that Fastly will use to encrypt your log files before writing them to disk"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	SecretKey         argparser.OptionalString    `help:"Your DigitalOcean Spaces account secret key"`
	TimestampFormat   argparser.OptionalString    `help:"strftime specified timestamp formatting (default \"%Y-%m-%dT%H:%M:%S.000\")"`

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
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateDigitalOceanInput, error) {
	input := fastly.UpdateDigitalOceanInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.BucketName.WasSet {
		input.BucketName = &c.BucketName.Value
	}

	if c.Domain.WasSet {
		input.Domain = &c.Domain.Value
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}

	if c.Period.WasSet {
		input.Period = &c.Period.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = &c.GzipLevel.Value
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

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		if strings.TrimSpace(c.Placement.Value) == "" {
			input.Placement = fastly.NullValue[string]()
		} else {
			input.Placement = fastly.NewNullable(c.Placement.Value)
		}
	}

	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
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

	digitalocean, err := c.Globals.APIClient.UpdateDigitalOcean(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Updated DigitalOcean Spaces logging endpoint %s (service %s version %d)",
		fastly.ToValue(digitalocean.Name),
		fastly.ToValue(digitalocean.ServiceID),
		fastly.ToValue(digitalocean.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *UpdateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "access-key", "autoclone", "bucket",
		"compression-codec", "domain", "format", "format-version",
		"gzip-level", "message-type", "new-name", "path", "period",
		"placement", "processing-region", "public-key", "response-condition",
		"secret-key", "service-id", "service-name", "timestamp-format",
	}
}
