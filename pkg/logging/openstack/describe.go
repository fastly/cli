package openstack

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an OpenStack logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetOpenstackInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an OpenStack logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the OpenStack logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	openstack, err := c.Globals.Client.GetOpenstack(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", openstack.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", openstack.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", openstack.Name)
	fmt.Fprintf(out, "Bucket: %s\n", openstack.BucketName)
	fmt.Fprintf(out, "Access key: %s\n", openstack.AccessKey)
	fmt.Fprintf(out, "User: %s\n", openstack.User)
	fmt.Fprintf(out, "URL: %s\n", openstack.URL)
	fmt.Fprintf(out, "Path: %s\n", openstack.Path)
	fmt.Fprintf(out, "Period: %d\n", openstack.Period)
	fmt.Fprintf(out, "GZip level: %d\n", openstack.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", openstack.Format)
	fmt.Fprintf(out, "Format version: %d\n", openstack.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", openstack.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", openstack.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", openstack.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", openstack.Placement)
	fmt.Fprintf(out, "Public key: %s\n", openstack.PublicKey)
	fmt.Fprintf(out, "Compression codec: %s\n", openstack.CompressionCodec)

	return nil
}
