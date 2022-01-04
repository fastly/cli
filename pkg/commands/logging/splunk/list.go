package splunk

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// ListCommand calls the Fastly API to list Splunk logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListSplunksInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Splunk endpoints on a Fastly service version")
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
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
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

	splunks, err := c.Globals.Client.ListSplunks(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, splunk := range splunks {
			tw.AddLine(splunk.ServiceID, splunk.ServiceVersion, splunk.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, splunk := range splunks {
		fmt.Fprintf(out, "\tSplunk %d/%d\n", i+1, len(splunks))
		fmt.Fprintf(out, "\t\tService ID: %s\n", splunk.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", splunk.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", splunk.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", splunk.URL)
		fmt.Fprintf(out, "\t\tToken: %s\n", splunk.Token)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", splunk.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", splunk.TLSHostname)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", splunk.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", splunk.TLSClientKey)
		fmt.Fprintf(out, "\t\tFormat: %s\n", splunk.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", splunk.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", splunk.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", splunk.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
