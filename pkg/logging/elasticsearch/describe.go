package elasticsearch

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an Elasticsearch logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetElasticsearchInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an Elasticsearch logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
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

	elasticsearch, err := c.Globals.Client.GetElasticsearch(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
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
