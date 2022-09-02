package syslog

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

// DescribeCommand calls the Fastly API to describe a Syslog logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSyslogInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a Syslog logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the Syslog logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	syslog, err := c.Globals.APIClient.GetSyslog(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(syslog)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", syslog.ServiceID)
	}
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
