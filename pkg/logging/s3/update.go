package s3

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Amazon S3 logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	Input fastly.GetS3Input

	NewName common.OptionalString

	BucketName common.OptionalString
	AccessKey  common.OptionalString
	SecretKey  common.OptionalString

	Domain                       common.OptionalString
	Path                         common.OptionalString
	Period                       common.OptionalUint
	GzipLevel                    common.OptionalUint
	Format                       common.OptionalString
	FormatVersion                common.OptionalUint
	MessageType                  common.OptionalString
	ResponseCondition            common.OptionalString
	TimestampFormat              common.OptionalString
	Redundancy                   common.OptionalString
	Placement                    common.OptionalString
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
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the S3 logging object").Short('n').Required().StringVar(&c.Input.Name)

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
	c.CmdClause.Flag("server-side-encryption", "Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms").Action(c.ServerSideEncryption.Set).EnumVar(&c.ServerSideEncryption.Value, string(fastly.S3ServerSideEncryptionAES), string(fastly.S3ServerSideEncryptionKMS))
	c.CmdClause.Flag("server-side-encryption-kms-key-id", "Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms").Action(c.ServerSideEncryptionKMSKeyID.Set).StringVar(&c.ServerSideEncryptionKMSKeyID.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	s3, err := c.Globals.Client.GetS3(&c.Input)
	if err != nil {
		return err
	}

	input := &fastly.UpdateS3Input{
		Service:                      s3.ServiceID,
		Version:                      s3.Version,
		Name:                         s3.Name,
		NewName:                      s3.Name,
		BucketName:                   s3.BucketName,
		Domain:                       s3.Domain,
		AccessKey:                    s3.AccessKey,
		SecretKey:                    s3.SecretKey,
		Path:                         s3.Path,
		Period:                       s3.Period,
		GzipLevel:                    s3.GzipLevel,
		Format:                       s3.Format,
		FormatVersion:                s3.FormatVersion,
		ResponseCondition:            s3.ResponseCondition,
		MessageType:                  s3.MessageType,
		TimestampFormat:              s3.TimestampFormat,
		Redundancy:                   s3.Redundancy,
		Placement:                    s3.Placement,
		ServerSideEncryption:         s3.ServerSideEncryption,
		ServerSideEncryptionKMSKeyID: s3.ServerSideEncryptionKMSKeyID,
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.BucketName.Valid {
		input.BucketName = c.BucketName.Value
	}

	if c.Domain.Valid {
		input.Domain = c.Domain.Value
	}

	if c.AccessKey.Valid {
		input.AccessKey = c.AccessKey.Value
	}

	if c.SecretKey.Valid {
		input.SecretKey = c.SecretKey.Value
	}

	if c.Path.Valid {
		input.Path = c.Path.Value
	}

	if c.Period.Valid {
		input.Period = c.Period.Value
	}

	if c.GzipLevel.Valid {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.Format.Valid {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.MessageType.Valid {
		input.MessageType = c.MessageType.Value
	}

	if c.TimestampFormat.Valid {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.Redundancy.Valid {
		switch c.Redundancy.Value {
		case string(fastly.S3RedundancyStandard):
			input.Redundancy = fastly.S3RedundancyStandard
		case string(fastly.S3RedundancyReduced):
			input.Redundancy = fastly.S3RedundancyReduced
		}
	}

	if c.Placement.Valid {
		input.Placement = c.Placement.Value
	}

	if c.ServerSideEncryption.Valid {
		switch c.ServerSideEncryption.Value {
		case string(fastly.S3ServerSideEncryptionAES):
			input.ServerSideEncryption = fastly.S3ServerSideEncryptionAES
		case string(fastly.S3ServerSideEncryptionKMS):
			input.ServerSideEncryption = fastly.S3ServerSideEncryptionKMS
		}
	}

	if c.ServerSideEncryptionKMSKeyID.Valid {
		input.ServerSideEncryptionKMSKeyID = c.ServerSideEncryptionKMSKeyID.Value
	}

	s3, err = c.Globals.Client.UpdateS3(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated S3 logging endpoint %s (service %s version %d)", s3.Name, s3.ServiceID, s3.Version)
	return nil
}
