package s3

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an Amazon S3 logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetS3Input
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a S3 logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the S3 logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	s3, err := c.Globals.Client.GetS3(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", s3.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", s3.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", s3.Name)
	fmt.Fprintf(out, "Bucket: %s\n", s3.BucketName)
	if s3.AccessKey != "" || s3.SecretKey != "" {
		fmt.Fprintf(out, "Access key: %s\n", s3.AccessKey)
		fmt.Fprintf(out, "Secret key: %s\n", s3.SecretKey)
	}
	if s3.IAMRole != "" {
		fmt.Fprintf(out, "IAM role: %s\n", s3.IAMRole)
	}
	fmt.Fprintf(out, "Path: %s\n", s3.Path)
	fmt.Fprintf(out, "Period: %d\n", s3.Period)
	fmt.Fprintf(out, "GZip level: %d\n", s3.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", s3.Format)
	fmt.Fprintf(out, "Format version: %d\n", s3.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", s3.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", s3.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", s3.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", s3.Placement)
	fmt.Fprintf(out, "Public key: %s\n", s3.PublicKey)
	fmt.Fprintf(out, "Redundancy: %s\n", s3.Redundancy)
	fmt.Fprintf(out, "Server-side encryption: %s\n", s3.ServerSideEncryption)
	fmt.Fprintf(out, "Server-side encryption KMS key ID: %s\n", s3.ServerSideEncryption)
	fmt.Fprintf(out, "Compression codec: %s\n", s3.CompressionCodec)

	return nil
}
