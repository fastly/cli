package s3

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Amazon S3 logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// mutual exclusions
	// AccessKey + SecretKey or IAMRole must be provided
	AccessKey argparser.OptionalString `help:"Your S3 account access key"`
	SecretKey argparser.OptionalString `help:"Your S3 account secret key"`
	IAMRole   argparser.OptionalString `help:"The IAM role ARN for logging"`

	// Optional.
	AutoClone                    argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	BucketName                   argparser.OptionalString    `name:"bucket" help:"Your S3 bucket name"`
	CompressionCodec             argparser.OptionalString    `help:"The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is \"gzip\", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error."`
	Domain                       argparser.OptionalString    `help:"The domain of the S3 endpoint"`
	EndpointName                 argparser.OptionalString    `name:"name" short:"n" help:"The name of the S3 logging object. Used as a primary key for API access"` // Can't shadow argparser.Base method Name().
	FileMaxBytes                 argparser.OptionalInt       `help:"The maximum size of a log file in bytes"`
	Format                       argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion                argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	GzipLevel                    argparser.OptionalInt       `help:"What level of GZIP encoding to have when dumping logs (default 0, no compression)"`
	MessageType                  argparser.OptionalString    `help:"How the message should be formatted. One of: classic (default), loggly, logplex or blank"`
	Path                         argparser.OptionalString    `help:"The path to upload logs to"`
	Period                       argparser.OptionalInt       `help:"How frequently log files are finalized so they can be available for reading (in seconds, default 3600)"`
	Placement                    argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion             argparser.OptionalString    `help:"The region where logs will be processed before streaming to S3. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	PublicKey                    argparser.OptionalString    `help:"A PGP public key that Fastly will use to encrypt your log files before writing them to disk"`
	Redundancy                   argparser.OptionalString    `hints:"standard,intelligent_tiering,standard_ia,onezone_ia,glacier,glacier_ir,deep_archive,reduced_redundancy" help:"The S3 storage class. One of: standard, intelligent_tiering, standard_ia, onezone_ia, glacier, glacier_ir, deep_archive, or reduced_redundancy"`
	ResponseCondition            argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	ServerSideEncryption         argparser.OptionalString    `hints:"AES256,aws:kms" help:"Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms"`
	ServerSideEncryptionKMSKeyID argparser.OptionalString    `help:"Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms"`
	TimestampFormat              argparser.OptionalString    `help:"strftime specified timestamp formatting (default \"%Y-%m-%dT%H:%M:%S.000\")"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Redundancy.Value != "" || kongcompat.Given(ctx, "redundancy")) && !slices.Contains([]string{"standard", "intelligent_tiering", "standard_ia", "onezone_ia", "glacier", "glacier_ir", "deep_archive", "reduced_redundancy"}, c.Redundancy.Value) {
		return fmt.Errorf("enum value must be one of standard,intelligent_tiering,standard_ia,onezone_ia,glacier,glacier_ir,deep_archive,reduced_redundancy, got '%s'", c.Redundancy.Value)
	}
	if (c.ServerSideEncryption.Value != "" || kongcompat.Given(ctx, "server-side-encryption")) && !slices.Contains([]string{"AES256", "aws:kms"}, c.ServerSideEncryption.Value) {
		return fmt.Errorf("enum value must be one of AES256,aws:kms, got '%s'", c.ServerSideEncryption.Value)
	}
	c.ServiceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateS3Input, error) {
	var input fastly.CreateS3Input

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.BucketName.WasSet {
		input.BucketName = &c.BucketName.Value
	}

	// The following block checks for invalid permutations of the ways in
	// which the AccessKey + SecretKey and IAMRole flags can be
	// provided. This is necessary because either the AccessKey and
	// SecretKey or the IAMRole is required, but they are mutually
	// exclusive. The kingpin library lacks a way to express this constraint
	// via the flag specification API so we enforce it manually here.
	switch {
	case !c.AccessKey.WasSet && !c.SecretKey.WasSet && !c.IAMRole.WasSet:
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags or the --iam-role flag must be provided")
	case (c.AccessKey.WasSet || c.SecretKey.WasSet) && c.IAMRole.WasSet:
		// Enforce mutual exclusion
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag")
	case c.AccessKey.WasSet && !c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --secret-key not provided")
	case !c.AccessKey.WasSet && c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --access-key not provided")
	}

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.IAMRole.WasSet {
		input.IAMRole = &c.IAMRole.Value
	}

	if c.Domain.WasSet {
		input.Domain = &c.Domain.Value
	}

	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = &c.FileMaxBytes.Value
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

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
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

	if c.ServerSideEncryptionKMSKeyID.WasSet {
		input.ServerSideEncryptionKMSKeyID = &c.ServerSideEncryptionKMSKeyID.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}

	if c.Redundancy.WasSet {
		redundancy, err := ValidateRedundancy(c.Redundancy.Value)
		if err == nil {
			input.Redundancy = &redundancy
		}
	}

	if c.ServerSideEncryption.WasSet {
		switch c.ServerSideEncryption.Value {
		case string(fastly.S3ServerSideEncryptionAES):
			sse := fastly.S3ServerSideEncryptionAES
			input.ServerSideEncryption = &sse
		case string(fastly.S3ServerSideEncryptionKMS):
			sse := fastly.S3ServerSideEncryptionKMS
			input.ServerSideEncryption = &sse
		}
	}

	return &input, nil
}

// ValidateRedundancy identifies the given redundancy type.
func ValidateRedundancy(val string) (redundancy fastly.S3Redundancy, err error) {
	switch val {
	case string(fastly.S3RedundancyStandard):
		redundancy = fastly.S3RedundancyStandard
	case string(fastly.S3RedundancyIntelligentTiering):
		redundancy = fastly.S3RedundancyIntelligentTiering
	case string(fastly.S3RedundancyStandardIA):
		redundancy = fastly.S3RedundancyStandardIA
	case string(fastly.S3RedundancyOneZoneIA):
		redundancy = fastly.S3RedundancyOneZoneIA
	case string(fastly.S3RedundancyGlacierInstantRetrieval):
		redundancy = fastly.S3RedundancyGlacierInstantRetrieval
	case string(fastly.S3RedundancyGlacierFlexibleRetrieval):
		redundancy = fastly.S3RedundancyGlacierFlexibleRetrieval
	case string(fastly.S3RedundancyGlacierDeepArchive):
		redundancy = fastly.S3RedundancyGlacierDeepArchive
	case string(fastly.S3RedundancyReduced):
		redundancy = fastly.S3RedundancyReduced
	default:
		err = fmt.Errorf("unknown redundancy: %s", val)
	}
	return redundancy, err
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

	d, err := c.Globals.APIClient.CreateS3(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Created S3 logging endpoint %s (service %s version %d)",
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
		"name", "version", "autoclone", "access-key", "bucket",
		"compression-codec", "domain", "file-max-bytes", "format",
		"format-version", "gzip-level", "iam-role", "message-type", "path",
		"period", "placement", "processing-region", "public-key", "redundancy",
		"response-condition", "secret-key", "server-side-encryption",
		"server-side-encryption-kms-key-id", "service-id", "service-name",
		"timestamp-format",
	}
}
