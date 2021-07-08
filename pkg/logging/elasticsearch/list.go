package elasticsearch

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Elasticsearch logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListElasticsearchInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Elasticsearch endpoints on a Fastly service version")
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

	elasticsearchs, err := c.Globals.Client.ListElasticsearch(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, elasticsearch := range elasticsearchs {
			tw.AddLine(elasticsearch.ServiceID, elasticsearch.ServiceVersion, elasticsearch.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, elasticsearch := range elasticsearchs {
		fmt.Fprintf(out, "\tElasticsearch %d/%d\n", i+1, len(elasticsearchs))
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
