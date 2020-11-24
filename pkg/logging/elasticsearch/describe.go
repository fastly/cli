package elasticsearch

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v2/fastly"
)

// DescribeCommand calls the Fastly API to describe an Elasticsearch logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetElasticsearchInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an Elasticsearch logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	elasticsearch, err := c.Globals.Client.GetElasticsearch(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", elasticsearch.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", elasticsearch.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", elasticsearch.Name)
	fmt.Fprintf(out, "Index: %s\n", elasticsearch.Index)
	fmt.Fprintf(out, "URL: %s\n", elasticsearch.URL)
	fmt.Fprintf(out, "Pipeline: %s\n", elasticsearch.Pipeline)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", elasticsearch.TLSCACert)
	fmt.Fprintf(out, "TLS client certificate: %s\n", elasticsearch.TLSClientCert)
	fmt.Fprintf(out, "TLS client key: %s\n", elasticsearch.TLSClientKey)
	fmt.Fprintf(out, "TLS hostname: %s\n", elasticsearch.TLSHostname)
	fmt.Fprintf(out, "User: %s\n", elasticsearch.User)
	fmt.Fprintf(out, "Password: %s\n", elasticsearch.Password)
	fmt.Fprintf(out, "Format: %s\n", elasticsearch.Format)
	fmt.Fprintf(out, "Format version: %d\n", elasticsearch.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", elasticsearch.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", elasticsearch.Placement)

	return nil
}
