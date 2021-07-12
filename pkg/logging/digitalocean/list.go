package digitalocean

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list DigitalOcean Spaces logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListDigitalOceansInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List DigitalOcean Spaces logging endpoints on a Fastly service version")
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

	digitaloceans, err := c.Globals.Client.ListDigitalOceans(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, digitalocean := range digitaloceans {
			tw.AddLine(digitalocean.ServiceID, digitalocean.ServiceVersion, digitalocean.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, digitalocean := range digitaloceans {
		fmt.Fprintf(out, "\tDigitalOcean %d/%d\n", i+1, len(digitaloceans))
		fmt.Fprintf(out, "\t\tService ID: %s\n", digitalocean.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", digitalocean.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", digitalocean.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", digitalocean.BucketName)
		fmt.Fprintf(out, "\t\tDomain: %s\n", digitalocean.Domain)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", digitalocean.AccessKey)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", digitalocean.SecretKey)
		fmt.Fprintf(out, "\t\tPath: %s\n", digitalocean.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", digitalocean.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", digitalocean.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", digitalocean.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", digitalocean.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", digitalocean.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", digitalocean.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", digitalocean.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", digitalocean.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", digitalocean.PublicKey)
	}
	fmt.Fprintln(out)

	return nil
}
