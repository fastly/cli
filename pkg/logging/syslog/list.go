package syslog

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Syslog logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListSyslogsInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Syslog endpoints on a Fastly service version")
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

	syslogs, err := c.Globals.Client.ListSyslogs(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, syslog := range syslogs {
			tw.AddLine(syslog.ServiceID, syslog.ServiceVersion, syslog.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, syslog := range syslogs {
		fmt.Fprintf(out, "\tSyslog %d/%d\n", i+1, len(syslogs))
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
