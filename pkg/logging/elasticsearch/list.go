package elasticsearch

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

// ListCommand calls the Fastly API to list Elasticsearch logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListElasticsearchInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Elasticsearch endpoints on a Fastly service version")
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

	elasticsearchs, err := c.Globals.Client.ListElasticsearch(&c.Input)
	if err != nil {
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
