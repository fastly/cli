package digitalocean

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe a DigitalOcean Spaces logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetDigitalOceanInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a DigitalOcean Spaces logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the DigitalOcean Spaces logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	digitalocean, err := c.Globals.Client.GetDigitalOcean(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", digitalocean.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", digitalocean.Version)
	fmt.Fprintf(out, "Name: %s\n", digitalocean.Name)
	fmt.Fprintf(out, "Bucket: %s\n", digitalocean.BucketName)
	fmt.Fprintf(out, "Domain: %s\n", digitalocean.Domain)
	fmt.Fprintf(out, "Access key: %s\n", digitalocean.AccessKey)
	fmt.Fprintf(out, "Secret key: %s\n", digitalocean.SecretKey)
	fmt.Fprintf(out, "Path: %s\n", digitalocean.Path)
	fmt.Fprintf(out, "Period: %d\n", digitalocean.Period)
	fmt.Fprintf(out, "GZip level: %d\n", digitalocean.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", digitalocean.Format)
	fmt.Fprintf(out, "Format version: %d\n", digitalocean.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", digitalocean.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", digitalocean.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", digitalocean.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", digitalocean.Placement)
	fmt.Fprintf(out, "Public key: %s\n", digitalocean.PublicKey)

	return nil
}
