package bigquery

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list BigQuery logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListBigQueriesInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List BigQuery endpoints on a Fastly service version")
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

	bqs, err := c.Globals.Client.ListBigQueries(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, bq := range bqs {
			tw.AddLine(bq.ServiceID, bq.ServiceVersion, bq.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, bq := range bqs {
		fmt.Fprintf(out, "\tBigQuery %d/%d\n", i+1, len(bqs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", bq.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", bq.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", bq.Name)
		fmt.Fprintf(out, "\t\tFormat: %s\n", bq.Format)
		fmt.Fprintf(out, "\t\tUser: %s\n", bq.User)
		fmt.Fprintf(out, "\t\tProject ID: %s\n", bq.ProjectID)
		fmt.Fprintf(out, "\t\tDataset: %s\n", bq.Dataset)
		fmt.Fprintf(out, "\t\tTable: %s\n", bq.Table)
		fmt.Fprintf(out, "\t\tTemplate suffix: %s\n", bq.Template)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", bq.SecretKey)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", bq.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", bq.Placement)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", bq.FormatVersion)
	}
	fmt.Fprintln(out)

	return nil
}
