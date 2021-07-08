package cloudfiles

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Cloudfiles logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetCloudfilesInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Cloudfiles logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Cloudfiles logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	cloudfiles, err := c.Globals.Client.GetCloudfiles(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", cloudfiles.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", cloudfiles.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", cloudfiles.Name)
	fmt.Fprintf(out, "User: %s\n", cloudfiles.User)
	fmt.Fprintf(out, "Access key: %s\n", cloudfiles.AccessKey)
	fmt.Fprintf(out, "Bucket: %s\n", cloudfiles.BucketName)
	fmt.Fprintf(out, "Path: %s\n", cloudfiles.Path)
	fmt.Fprintf(out, "Region: %s\n", cloudfiles.Region)
	fmt.Fprintf(out, "Placement: %s\n", cloudfiles.Placement)
	fmt.Fprintf(out, "Period: %d\n", cloudfiles.Period)
	fmt.Fprintf(out, "GZip level: %d\n", cloudfiles.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", cloudfiles.Format)
	fmt.Fprintf(out, "Format version: %d\n", cloudfiles.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", cloudfiles.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", cloudfiles.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", cloudfiles.TimestampFormat)
	fmt.Fprintf(out, "Public key: %s\n", cloudfiles.PublicKey)

	return nil
}
