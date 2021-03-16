package kinesis

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an Amazon Kinesis logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetKinesisInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Kinesis logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Kinesis logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	kinesis, err := c.Globals.Client.GetKinesis(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", kinesis.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", kinesis.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", kinesis.Name)
	fmt.Fprintf(out, "Stream name: %s\n", kinesis.StreamName)
	fmt.Fprintf(out, "Region: %s\n", kinesis.Region)
	fmt.Fprintf(out, "Access key: %s\n", kinesis.AccessKey)
	fmt.Fprintf(out, "Secret key: %s\n", kinesis.SecretKey)
	fmt.Fprintf(out, "Format: %s\n", kinesis.Format)
	fmt.Fprintf(out, "Format version: %d\n", kinesis.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", kinesis.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", kinesis.Placement)

	return nil
}
