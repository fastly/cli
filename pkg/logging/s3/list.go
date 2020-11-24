package s3

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// ListCommand calls the Fastly API to list Amazon S3 logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListS3sInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List S3 endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	s3s, err := c.Globals.Client.ListS3s(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, s3 := range s3s {
			tw.AddLine(s3.ServiceID, s3.ServiceVersion, s3.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, s3 := range s3s {
		fmt.Fprintf(out, "\tS3 %d/%d\n", i+1, len(s3s))
		fmt.Fprintf(out, "\t\tService ID: %s\n", s3.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", s3.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", s3.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", s3.BucketName)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", s3.AccessKey)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", s3.SecretKey)
		fmt.Fprintf(out, "\t\tPath: %s\n", s3.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", s3.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", s3.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", s3.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", s3.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", s3.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", s3.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", s3.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", s3.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", s3.PublicKey)
		fmt.Fprintf(out, "\t\tRedundancy: %s\n", s3.Redundancy)
		fmt.Fprintf(out, "\t\tServer-side encryption: %s\n", s3.ServerSideEncryption)
		fmt.Fprintf(out, "\t\tServer-side encryption KMS key ID: %s\n", s3.ServerSideEncryption)
	}
	fmt.Fprintln(out)

	return nil
}
