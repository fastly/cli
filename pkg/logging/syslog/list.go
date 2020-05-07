package syslog

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list Syslog logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListSyslogsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Syslog endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	syslogs, err := c.Globals.Client.ListSyslogs(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, syslog := range syslogs {
			tw.AddLine(syslog.ServiceID, syslog.Version, syslog.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, syslog := range syslogs {
		fmt.Fprintf(out, "\tSyslog %d/%d\n", i+1, len(syslogs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", syslog.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", syslog.Version)
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
