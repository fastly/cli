package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v6/fastly"
)

// DescribeCommand calls the Fastly API to describe an Elasticsearch logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetElasticsearchInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about an Elasticsearch logging endpoint on a Fastly service version").Alias("get")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	elasticsearch, err := c.Globals.APIClient.GetElasticsearch(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(elasticsearch)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", elasticsearch.ServiceID)
	}
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
