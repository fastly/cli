package splunk

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v5/fastly"
)

// DescribeCommand calls the Fastly API to describe a Splunk logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSplunkInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a Splunk logging endpoint on a Fastly service version").Alias("get")
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
	})
	c.CmdClause.Flag("name", "The name of the Splunk logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	splunk, err := c.Globals.Client.GetSplunk(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", splunk.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", splunk.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", splunk.Name)
	fmt.Fprintf(out, "URL: %s\n", splunk.URL)
	fmt.Fprintf(out, "Token: %s\n", splunk.Token)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", splunk.TLSCACert)
	fmt.Fprintf(out, "TLS hostname: %s\n", splunk.TLSHostname)
	fmt.Fprintf(out, "TLS client certificate: %s\n", splunk.TLSClientCert)
	fmt.Fprintf(out, "TLS client key: %s\n", splunk.TLSClientKey)
	fmt.Fprintf(out, "Format: %s\n", splunk.Format)
	fmt.Fprintf(out, "Format version: %d\n", splunk.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", splunk.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", splunk.Placement)

	return nil
}
