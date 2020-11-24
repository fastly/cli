package bigquery

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v2/fastly"
)

// DescribeCommand calls the Fastly API to describe a BigQuery logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetBigQueryInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a BigQuery logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the BigQuery logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	bq, err := c.Globals.Client.GetBigQuery(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", bq.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", bq.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", bq.Name)
	fmt.Fprintf(out, "Format: %s\n", bq.Format)
	fmt.Fprintf(out, "User: %s\n", bq.User)
	fmt.Fprintf(out, "Project ID: %s\n", bq.ProjectID)
	fmt.Fprintf(out, "Dataset: %s\n", bq.Dataset)
	fmt.Fprintf(out, "Table: %s\n", bq.Table)
	fmt.Fprintf(out, "Template suffix: %s\n", bq.Template)
	fmt.Fprintf(out, "Secret key: %s\n", bq.SecretKey)
	fmt.Fprintf(out, "Response condition: %s\n", bq.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", bq.Placement)
	fmt.Fprintf(out, "Format version: %d\n", bq.FormatVersion)

	return nil
}
