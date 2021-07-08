package azureblob

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Azure Blob Storage logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListBlobStoragesInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Azure Blob Storage logging endpoints on a Fastly service version")
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

	azureblobs, err := c.Globals.Client.ListBlobStorages(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, azureblob := range azureblobs {
			tw.AddLine(azureblob.ServiceID, azureblob.ServiceVersion, azureblob.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, azureblob := range azureblobs {
		fmt.Fprintf(out, "\tBlobStorage %d/%d\n", i+1, len(azureblobs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", azureblob.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", azureblob.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", azureblob.Name)
		fmt.Fprintf(out, "\t\tContainer: %s\n", azureblob.Container)
		fmt.Fprintf(out, "\t\tAccount name: %s\n", azureblob.AccountName)
		fmt.Fprintf(out, "\t\tSAS token: %s\n", azureblob.SASToken)
		fmt.Fprintf(out, "\t\tPath: %s\n", azureblob.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", azureblob.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", azureblob.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", azureblob.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", azureblob.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", azureblob.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", azureblob.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", azureblob.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", azureblob.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", azureblob.PublicKey)
		fmt.Fprintf(out, "\t\tFile max bytes: %d\n", azureblob.FileMaxBytes)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", azureblob.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
