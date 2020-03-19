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

// CreateCommand calls the Fastly API to create Amazon S3 logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateS3Input

	redundancy           string
	serverSideEncryption string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an Amazon S3 logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the S3 logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)

	c.CmdClause.Flag("bucket", "Your S3 bucket name").Required().StringVar(&c.Input.BucketName)
	c.CmdClause.Flag("access-key", "Your S3 account access key").Required().StringVar(&c.Input.AccessKey)
	c.CmdClause.Flag("secret-key", "Your S3 account secret key").Required().StringVar(&c.Input.SecretKey)

	c.CmdClause.Flag("domain", "The domain of the S3 endpoint").StringVar(&c.Input.Domain)
	c.CmdClause.Flag("path", "The path to upload logs to").StringVar(&c.Input.Path)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").UintVar(&c.Input.Period)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").UintVar(&c.Input.GzipLevel)
	c.CmdClause.Flag("format", "Apache style log formatting").StringVar(&c.Input.Format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").UintVar(&c.Input.FormatVersion)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").StringVar(&c.Input.MessageType)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").StringVar(&c.Input.ResponseCondition)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).StringVar(&c.Input.TimestampFormat)
	c.CmdClause.Flag("redundancy", "The S3 redundancy level. Can be either standard or reduced_redundancy").EnumVar(&c.redundancy, string(fastly.S3RedundancyStandard), string(fastly.S3RedundancyReduced))
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").StringVar(&c.Input.Placement)
	c.CmdClause.Flag("server-side-encryption", "Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms").EnumVar(&c.serverSideEncryption, string(fastly.S3ServerSideEncryptionAES), string(fastly.S3ServerSideEncryptionKMS))
	c.CmdClause.Flag("server-side-encryption-kms-key-id", "Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms").StringVar(&c.Input.ServerSideEncryptionKMSKeyID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	switch c.redundancy {
	case string(fastly.S3RedundancyStandard):
		c.Input.Redundancy = fastly.S3RedundancyStandard
	case string(fastly.S3RedundancyReduced):
		c.Input.Redundancy = fastly.S3RedundancyReduced
	}

	switch c.serverSideEncryption {
	case string(fastly.S3ServerSideEncryptionAES):
		c.Input.ServerSideEncryption = fastly.S3ServerSideEncryptionAES
	case string(fastly.S3ServerSideEncryptionKMS):
		c.Input.ServerSideEncryption = fastly.S3ServerSideEncryptionKMS
	}

	d, err := c.Globals.Client.CreateS3(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created S3 logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
