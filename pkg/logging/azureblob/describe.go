package azureblob

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an Azure Blob Storage logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetBlobStorageInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an Azure Blob Storage logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	azureblob, err := c.Globals.Client.GetBlobStorage(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", azureblob.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", azureblob.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", azureblob.Name)
	fmt.Fprintf(out, "Container: %s\n", azureblob.Container)
	fmt.Fprintf(out, "Account name: %s\n", azureblob.AccountName)
	fmt.Fprintf(out, "SAS token: %s\n", azureblob.SASToken)
	fmt.Fprintf(out, "Path: %s\n", azureblob.Path)
	fmt.Fprintf(out, "Period: %d\n", azureblob.Period)
	fmt.Fprintf(out, "GZip level: %d\n", azureblob.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", azureblob.Format)
	fmt.Fprintf(out, "Format version: %d\n", azureblob.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", azureblob.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", azureblob.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", azureblob.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", azureblob.Placement)
	fmt.Fprintf(out, "Public key: %s\n", azureblob.PublicKey)

	return nil
}
