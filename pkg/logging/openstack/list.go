package openstack

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list OpenStack logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListOpenstackInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List OpenStack logging endpoints on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
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

	openstacks, err := c.Globals.Client.ListOpenstack(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
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
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", openstack.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
