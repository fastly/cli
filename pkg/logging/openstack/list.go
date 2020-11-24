package openstack

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

// ListCommand calls the Fastly API to list OpenStack logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListOpenstackInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List OpenStack logging endpoints on a Fastly service version")
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

	openstacks, err := c.Globals.Client.ListOpenstack(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, openstack := range openstacks {
			tw.AddLine(openstack.ServiceID, openstack.ServiceVersion, openstack.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, openstack := range openstacks {
		fmt.Fprintf(out, "\tOpenstack %d/%d\n", i+1, len(openstacks))
		fmt.Fprintf(out, "\t\tService ID: %s\n", openstack.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", openstack.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", openstack.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", openstack.BucketName)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", openstack.AccessKey)
		fmt.Fprintf(out, "\t\tUser: %s\n", openstack.User)
		fmt.Fprintf(out, "\t\tURL: %s\n", openstack.URL)
		fmt.Fprintf(out, "\t\tPath: %s\n", openstack.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", openstack.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", openstack.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", openstack.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", openstack.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", openstack.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", openstack.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", openstack.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", openstack.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", openstack.PublicKey)
	}
	fmt.Fprintln(out)

	return nil
}
