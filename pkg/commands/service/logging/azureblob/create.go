package azureblob

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Azure Blob Storage logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// Optional.
	EndpointName      argparser.OptionalString    `name:"name" short:"n" help:"The name of the Azure Blob Storage logging object. Used as a primary key for API access"`
	Container         argparser.OptionalString    `help:"The name of the Azure Blob Storage container in which to store logs"`
	AccountName       argparser.OptionalString    `help:"The unique Azure Blob Storage namespace in which your data objects are stored"`
	SASToken          argparser.OptionalString    `help:"The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work"`
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Path              argparser.OptionalString    `help:"The path to upload logs to"`
	Period            argparser.OptionalInt       `help:"How frequently log files are finalized so they can be available for reading (in seconds, default 3600)"`
	GzipLevel         argparser.OptionalInt       `help:"What level of GZIP encoding to have when dumping logs (default 0, no compression)"`
	MessageType       argparser.OptionalString    `help:"How the message should be formatted. One of: classic (default), loggly, logplex or blank"`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	TimestampFormat   argparser.OptionalString    `help:"strftime specified timestamp formatting (default \"%Y-%m-%dT%H:%M:%S.000\")"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to Azure Blob Storage. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	PublicKey         argparser.OptionalString    `help:"A PGP public key that Fastly will use to encrypt your log files before writing them to disk"`
	FileMaxBytes      argparser.OptionalInt       `help:"The maximum size of a log file in bytes"`
	CompressionCodec  argparser.OptionalString    `help:"The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is \"gzip\", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error."`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.ServiceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateBlobStorageInput, error) {
	var input fastly.CreateBlobStorageInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Container.WasSet {
		input.Container = &c.Container.Value
	}
	if c.AccountName.WasSet {
		input.AccountName = &c.AccountName.Value
	}
	if c.SASToken.WasSet {
		input.SASToken = &c.SASToken.Value
	}

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
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
		input.Placement = &c.Placement.Value
	}
	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}
	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}
	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = &c.FileMaxBytes.Value
	}
	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	d, err := c.Globals.APIClient.CreateBlobStorage(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(
		out,
		"Created Azure Blob Storage logging endpoint %s (service %s version %d)",
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
		"version", "account-name", "autoclone", "compression-codec",
		"container", "file-max-bytes", "format", "format-version",
		"gzip-level", "message-type", "name", "path", "period", "placement",
		"processing-region", "public-key", "response-condition", "sas-token",
		"service-id", "service-name", "timestamp-format",
	}
}
