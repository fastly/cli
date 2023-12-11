package elasticsearch

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
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
	c.Input.ServiceVersion = serviceVersion.Number

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
			tw.AddLine(elasticsearch.ServiceID, elasticsearch.ServiceVersion, elasticsearch.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, elasticsearch := range o {
		fmt.Fprintf(out, "\tElasticsearch %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", elasticsearch.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", elasticsearch.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", elasticsearch.Name)
		fmt.Fprintf(out, "\t\tIndex: %s\n", elasticsearch.Index)
		fmt.Fprintf(out, "\t\tURL: %s\n", elasticsearch.URL)
		fmt.Fprintf(out, "\t\tPipeline: %s\n", elasticsearch.Pipeline)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", elasticsearch.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", elasticsearch.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", elasticsearch.TLSClientKey)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", elasticsearch.TLSHostname)
		fmt.Fprintf(out, "\t\tUser: %s\n", elasticsearch.User)
		fmt.Fprintf(out, "\t\tPassword: %s\n", elasticsearch.Password)
		fmt.Fprintf(out, "\t\tFormat: %s\n", elasticsearch.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", elasticsearch.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", elasticsearch.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", elasticsearch.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
