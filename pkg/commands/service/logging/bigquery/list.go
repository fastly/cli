package bigquery

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list BigQuery logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListBigQueriesInput      `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListBigQueries(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, bq := range o {
			tw.AddLine(
				fastly.ToValue(bq.ServiceID),
				fastly.ToValue(bq.ServiceVersion),
				fastly.ToValue(bq.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, bq := range o {
		fmt.Fprintf(out, "\tBigQuery %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(bq.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(bq.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(bq.Name))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(bq.Format))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(bq.User))
		fmt.Fprintf(out, "\t\tAccount name: %s\n", fastly.ToValue(bq.AccountName))
		fmt.Fprintf(out, "\t\tProject ID: %s\n", fastly.ToValue(bq.ProjectID))
		fmt.Fprintf(out, "\t\tDataset: %s\n", fastly.ToValue(bq.Dataset))
		fmt.Fprintf(out, "\t\tTable: %s\n", fastly.ToValue(bq.Table))
		fmt.Fprintf(out, "\t\tTemplate suffix: %s\n", fastly.ToValue(bq.Template))
		fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(bq.SecretKey))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(bq.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(bq.Placement))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(bq.FormatVersion))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(bq.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
