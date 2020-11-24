package azureblob

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

// ListCommand calls the Fastly API to list Azure Blob Storage logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListBlobStoragesInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Azure Blob Storage logging endpoints on a Fastly service version")
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

	azureblobs, err := c.Globals.Client.ListBlobStorages(&c.Input)
	if err != nil {
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
	}
	fmt.Fprintln(out)

	return nil
}
