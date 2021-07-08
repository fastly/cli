package gcs

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list GCS logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListGCSsInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List GCS endpoints on a Fastly service version")
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

	gcss, err := c.Globals.Client.ListGCSs(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, gcs := range gcss {
			tw.AddLine(gcs.ServiceID, gcs.ServiceVersion, gcs.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, gcs := range gcss {
		fmt.Fprintf(out, "\tGCS %d/%d\n", i+1, len(gcss))
		fmt.Fprintf(out, "\t\tService ID: %s\n", gcs.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", gcs.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", gcs.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", gcs.Bucket)
		fmt.Fprintf(out, "\t\tUser: %s\n", gcs.User)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", gcs.SecretKey)
		fmt.Fprintf(out, "\t\tPath: %s\n", gcs.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", gcs.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", gcs.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", gcs.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", gcs.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", gcs.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", gcs.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", gcs.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", gcs.Placement)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", gcs.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
