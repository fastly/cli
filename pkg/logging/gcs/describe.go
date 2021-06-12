package gcs

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a GCS logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.GetGCSInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a GCS logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Envar(env.ServiceID).StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the GCS logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	gcs, err := c.Globals.Client.GetGCS(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", gcs.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", gcs.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", gcs.Name)
	fmt.Fprintf(out, "Bucket: %s\n", gcs.Bucket)
	fmt.Fprintf(out, "User: %s\n", gcs.User)
	fmt.Fprintf(out, "Secret key: %s\n", gcs.SecretKey)
	fmt.Fprintf(out, "Path: %s\n", gcs.Path)
	fmt.Fprintf(out, "Period: %d\n", gcs.Period)
	fmt.Fprintf(out, "GZip level: %d\n", gcs.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", gcs.Format)
	fmt.Fprintf(out, "Format version: %d\n", gcs.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", gcs.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", gcs.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", gcs.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", gcs.Placement)
	fmt.Fprintf(out, "Compression codec: %s\n", gcs.CompressionCodec)

	return nil
}
