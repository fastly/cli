package syslog

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Syslog logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSyslogInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Syslog logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Syslog logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	syslog, err := c.Globals.Client.GetSyslog(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", syslog.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", syslog.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", syslog.Name)
	fmt.Fprintf(out, "Address: %s\n", syslog.Address)
	fmt.Fprintf(out, "Hostname: %s\n", syslog.Hostname)
	fmt.Fprintf(out, "Port: %d\n", syslog.Port)
	fmt.Fprintf(out, "Use TLS: %t\n", syslog.UseTLS)
	fmt.Fprintf(out, "IPV4: %s\n", syslog.IPV4)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", syslog.TLSCACert)
	fmt.Fprintf(out, "TLS hostname: %s\n", syslog.TLSHostname)
	fmt.Fprintf(out, "TLS client certificate: %s\n", syslog.TLSClientCert)
	fmt.Fprintf(out, "TLS client key: %s\n", syslog.TLSClientKey)
	fmt.Fprintf(out, "Token: %s\n", syslog.Token)
	fmt.Fprintf(out, "Format: %s\n", syslog.Format)
	fmt.Fprintf(out, "Format version: %d\n", syslog.FormatVersion)
	fmt.Fprintf(out, "Message type: %s\n", syslog.MessageType)
	fmt.Fprintf(out, "Response condition: %s\n", syslog.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", syslog.Placement)

	return nil
}
