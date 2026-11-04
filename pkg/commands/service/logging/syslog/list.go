package syslog

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Syslog logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListSyslogsInput         `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
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
		ServiceNameFlag:    c.ServiceName,
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

	o, err := c.Globals.APIClient.ListSyslogs(context.TODO(), &c.Input)
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
		for _, syslog := range o {
			tw.AddLine(
				fastly.ToValue(syslog.ServiceID),
				fastly.ToValue(syslog.ServiceVersion),
				fastly.ToValue(syslog.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, syslog := range o {
		fmt.Fprintf(out, "\tSyslog %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(syslog.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(syslog.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(syslog.Name))
		fmt.Fprintf(out, "\t\tAddress: %s\n", fastly.ToValue(syslog.Address))
		fmt.Fprintf(out, "\t\tHostname: %s\n", fastly.ToValue(syslog.Hostname))
		fmt.Fprintf(out, "\t\tPort: %d\n", fastly.ToValue(syslog.Port))
		fmt.Fprintf(out, "\t\tUse TLS: %t\n", fastly.ToValue(syslog.UseTLS))
		fmt.Fprintf(out, "\t\tIPV4: %s\n", fastly.ToValue(syslog.IPV4))
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", fastly.ToValue(syslog.TLSCACert))
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", fastly.ToValue(syslog.TLSHostname))
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", fastly.ToValue(syslog.TLSClientCert))
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", fastly.ToValue(syslog.TLSClientKey))
		fmt.Fprintf(out, "\t\tToken: %s\n", fastly.ToValue(syslog.Token))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(syslog.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(syslog.FormatVersion))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(syslog.MessageType))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(syslog.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(syslog.Placement))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(syslog.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
