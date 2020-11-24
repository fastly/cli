package s3

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update Amazon S3 logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName                      common.OptionalString
	Address                      common.OptionalString
	BucketName                   common.OptionalString
	AccessKey                    common.OptionalString
	SecretKey                    common.OptionalString
	Domain                       common.OptionalString
	Path                         common.OptionalString
	Period                       common.OptionalUint
	GzipLevel                    common.OptionalUint
	Format                       common.OptionalString
	FormatVersion                common.OptionalUint
	MessageType                  common.OptionalString
	ResponseCondition            common.OptionalString
	TimestampFormat              common.OptionalString
	Placement                    common.OptionalString
	PublicKey                    common.OptionalString
	Redundancy                   common.OptionalString
	ServerSideEncryption         common.OptionalString
	ServerSideEncryptionKMSKeyID common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a S3 logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the S3 logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the S3 logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("bucket", "Your S3 bucket name").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	c.CmdClause.Flag("access-key", "Your S3 account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("secret-key", "Your S3 account secret key").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("domain", "The domain of the S3 endpoint").Action(c.Domain.Set).StringVar(&c.Domain.Value)
	c.CmdClause.Flag("path", "The path to upload logs to").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).UintVar(&c.GzipLevel.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("redundancy", "The S3 redundancy level. Can be either standard or reduced_redundancy").Action(c.Redundancy.Set).EnumVar(&c.Redundancy.Value, string(fastly.S3RedundancyStandard), string(fastly.S3RedundancyReduced))
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.PublicKey.Set).StringVar(&c.PublicKey.Value)
	c.CmdClause.Flag("server-side-encryption", "Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms").Action(c.ServerSideEncryption.Set).EnumVar(&c.ServerSideEncryption.Value, string(fastly.S3ServerSideEncryptionAES), string(fastly.S3ServerSideEncryptionKMS))
	c.CmdClause.Flag("server-side-encryption-kms-key-id", "Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms").Action(c.ServerSideEncryptionKMSKeyID.Set).StringVar(&c.ServerSideEncryptionKMSKeyID.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateS3Input, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	s3, err := c.Globals.Client.GetS3(&fastly.GetS3Input{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateS3Input{
		ServiceID:                    s3.ServiceID,
		ServiceVersion:               s3.ServiceVersion,
		Name:                         s3.Name,
		NewName:                      fastly.String(s3.Name),
		BucketName:                   fastly.String(s3.BucketName),
		Domain:                       fastly.String(s3.Domain),
		AccessKey:                    fastly.String(s3.AccessKey),
		SecretKey:                    fastly.String(s3.SecretKey),
		Path:                         fastly.String(s3.Path),
		Period:                       fastly.Uint(s3.Period),
		GzipLevel:                    fastly.Uint(s3.GzipLevel),
		Format:                       fastly.String(s3.Format),
		FormatVersion:                fastly.Uint(s3.FormatVersion),
		ResponseCondition:            fastly.String(s3.ResponseCondition),
		MessageType:                  fastly.String(s3.MessageType),
		TimestampFormat:              fastly.String(s3.TimestampFormat),
		Redundancy:                   s3.Redundancy,
		Placement:                    fastly.String(s3.Placement),
		PublicKey:                    fastly.String(s3.PublicKey),
		ServerSideEncryption:         s3.ServerSideEncryption,
		ServerSideEncryptionKMSKeyID: fastly.String(s3.ServerSideEncryptionKMSKeyID),
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.BucketName.WasSet {
		input.BucketName = fastly.String(c.BucketName.Value)
	}

	if c.AccessKey.WasSet {
		input.AccessKey = fastly.String(c.AccessKey.Value)
	}

	if c.SecretKey.WasSet {
		input.SecretKey = fastly.String(c.SecretKey.Value)
	}

	if c.Domain.WasSet {
		input.Domain = fastly.String(c.Domain.Value)
	}

	if c.Path.WasSet {
		input.Path = fastly.String(c.Path.Value)
	}

	if c.Period.WasSet {
		input.Period = fastly.Uint(c.Period.Value)
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = fastly.Uint(c.GzipLevel.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.MessageType.WasSet {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = fastly.String(c.TimestampFormat.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.PublicKey.WasSet {
		input.PublicKey = fastly.String(c.PublicKey.Value)
	}

	if c.ServerSideEncryptionKMSKeyID.WasSet {
		input.ServerSideEncryptionKMSKeyID = fastly.String(c.ServerSideEncryptionKMSKeyID.Value)
	}

	if c.Redundancy.WasSet {
		switch c.Redundancy.Value {
		case string(fastly.S3RedundancyStandard):
			input.Redundancy = fastly.S3RedundancyStandard
		case string(fastly.S3RedundancyReduced):
			input.Redundancy = fastly.S3RedundancyReduced
		}
	}

	if c.ServerSideEncryption.WasSet {
		switch c.ServerSideEncryption.Value {
		case string(fastly.S3ServerSideEncryptionAES):
			input.ServerSideEncryption = fastly.S3ServerSideEncryptionAES
		case string(fastly.S3ServerSideEncryptionKMS):
			input.ServerSideEncryption = fastly.S3ServerSideEncryptionKMS
		}
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	s3, err := c.Globals.Client.UpdateS3(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated S3 logging endpoint %s (service %s version %d)", s3.Name, s3.ServiceID, s3.ServiceVersion)
	return nil
}
