package elasticsearch

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// ListCommand calls the Fastly API to list Elasticsearch logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListElasticsearchInput
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
	c.CmdClause = parent.Command("list", "List Elasticsearch endpoints on a Fastly service version")

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
		Description: argparser.FlagServiceDesc,
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
		AllowActiveLocked:  true,
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

	o, err := c.Globals.APIClient.ListElasticsearch(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, elasticsearch := range o {
			tw.AddLine(
				fastly.ToValue(elasticsearch.ServiceID),
				fastly.ToValue(elasticsearch.ServiceVersion),
				fastly.ToValue(elasticsearch.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, elasticsearch := range o {
		fmt.Fprintf(out, "\tElasticsearch %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(elasticsearch.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(elasticsearch.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(elasticsearch.Name))
		fmt.Fprintf(out, "\t\tIndex: %s\n", fastly.ToValue(elasticsearch.Index))
		fmt.Fprintf(out, "\t\tURL: %s\n", fastly.ToValue(elasticsearch.URL))
		fmt.Fprintf(out, "\t\tPipeline: %s\n", fastly.ToValue(elasticsearch.Pipeline))
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", fastly.ToValue(elasticsearch.TLSCACert))
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", fastly.ToValue(elasticsearch.TLSClientCert))
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", fastly.ToValue(elasticsearch.TLSClientKey))
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", fastly.ToValue(elasticsearch.TLSHostname))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(elasticsearch.User))
		fmt.Fprintf(out, "\t\tPassword: %s\n", fastly.ToValue(elasticsearch.Password))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(elasticsearch.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(elasticsearch.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(elasticsearch.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(elasticsearch.Placement))
	}
	fmt.Fprintln(out)

	return nil
}
