package bigquery

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list BigQuery logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListBigQueriesInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List BigQuery endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
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
		ServiceNameFlag:    c.serviceName,
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

	o, err := c.Globals.APIClient.ListBigQueries(&c.Input)
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
	}
	fmt.Fprintln(out)

	return nil
}
