package syslog

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Syslog logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.ListSyslogsInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Syslog endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
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

	o, err := c.Globals.APIClient.ListSyslogs(&c.Input)
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
			tw.AddLine(syslog.ServiceID, syslog.ServiceVersion, syslog.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, syslog := range o {
		fmt.Fprintf(out, "\tSyslog %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", syslog.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", syslog.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", syslog.Name)
		fmt.Fprintf(out, "\t\tAddress: %s\n", syslog.Address)
		fmt.Fprintf(out, "\t\tHostname: %s\n", syslog.Hostname)
		fmt.Fprintf(out, "\t\tPort: %d\n", syslog.Port)
		fmt.Fprintf(out, "\t\tUse TLS: %t\n", syslog.UseTLS)
		fmt.Fprintf(out, "\t\tIPV4: %s\n", syslog.IPV4)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", syslog.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", syslog.TLSHostname)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", syslog.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", syslog.TLSClientKey)
		fmt.Fprintf(out, "\t\tToken: %s\n", syslog.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", syslog.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", syslog.FormatVersion)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", syslog.MessageType)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", syslog.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", syslog.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
