package openstack

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe an OpenStack logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetOpenstackInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an OpenStack logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the OpenStack logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	openstack, err := c.Globals.Client.GetOpenstack(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", openstack.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", openstack.Version)
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

	return nil
}
