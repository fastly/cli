package cloudfiles

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Cloudfiles logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListCloudfilesInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Cloudfiles endpoints on a Fastly service version")
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

	cloudfiles, err := c.Globals.Client.ListCloudfiles(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, cloudfile := range cloudfiles {
			tw.AddLine(cloudfile.ServiceID, cloudfile.ServiceVersion, cloudfile.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, cloudfile := range cloudfiles {
		fmt.Fprintf(out, "\tCloudfiles %d/%d\n", i+1, len(cloudfiles))
		fmt.Fprintf(out, "\t\tService ID: %s\n", cloudfile.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", cloudfile.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", cloudfile.Name)
		fmt.Fprintf(out, "\t\tUser: %s\n", cloudfile.User)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", cloudfile.AccessKey)
		fmt.Fprintf(out, "\t\tBucket: %s\n", cloudfile.BucketName)
		fmt.Fprintf(out, "\t\tPath: %s\n", cloudfile.Path)
		fmt.Fprintf(out, "\t\tRegion: %s\n", cloudfile.Region)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", cloudfile.Placement)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", cloudfile.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", cloudfile.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", cloudfile.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", cloudfile.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", cloudfile.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", cloudfile.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", cloudfile.TimestampFormat)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", cloudfile.PublicKey)
	}
	fmt.Fprintln(out)

	return nil
}
