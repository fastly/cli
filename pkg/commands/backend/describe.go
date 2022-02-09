package backend

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

// DescribeCommand calls the Fastly API to describe a backend.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetBackendInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a backend on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "Name of backend").Short('n').Required().StringVar(&c.Input.Name)
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

	backend, err := c.Globals.APIClient.GetBackend(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	err = c.print(out, backend)
	if err != nil {
		return err
	}
	return nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, b *fastly.Backend) error {
	if c.json {
		data, err := json.Marshal(b)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", b.ServiceID)
	}
	fmt.Fprintf(out, "Service Version: %d\n\n", b.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", b.Name)
	fmt.Fprintf(out, "Comment: %v\n", b.Comment)
	fmt.Fprintf(out, "Address: %v\n", b.Address)
	fmt.Fprintf(out, "Port: %v\n", b.Port)
	fmt.Fprintf(out, "Override host: %v\n", b.OverrideHost)
	fmt.Fprintf(out, "Connect timeout: %v\n", b.ConnectTimeout)
	fmt.Fprintf(out, "Max connections: %v\n", b.MaxConn)
	fmt.Fprintf(out, "First byte timeout: %v\n", b.FirstByteTimeout)
	fmt.Fprintf(out, "Between bytes timeout: %v\n", b.BetweenBytesTimeout)
	fmt.Fprintf(out, "Auto loadbalance: %v\n", b.AutoLoadbalance)
	fmt.Fprintf(out, "Weight: %v\n", b.Weight)
	fmt.Fprintf(out, "Healthcheck: %v\n", b.HealthCheck)
	fmt.Fprintf(out, "Shield: %v\n", b.Shield)
	fmt.Fprintf(out, "Use SSL: %v\n", b.UseSSL)
	fmt.Fprintf(out, "SSL check cert: %v\n", b.SSLCheckCert)
	fmt.Fprintf(out, "SSL CA cert: %v\n", b.SSLCACert)
	fmt.Fprintf(out, "SSL client cert: %v\n", b.SSLClientCert)
	fmt.Fprintf(out, "SSL client key: %v\n", b.SSLClientKey)
	fmt.Fprintf(out, "SSL cert hostname: %v\n", b.SSLCertHostname)
	fmt.Fprintf(out, "SSL SNI hostname: %v\n", b.SSLSNIHostname)
	fmt.Fprintf(out, "Min TLS version: %v\n", b.MinTLSVersion)
	fmt.Fprintf(out, "Max TLS version: %v\n", b.MaxTLSVersion)
	fmt.Fprintf(out, "SSL ciphers: %v\n", b.SSLCiphers)

	return nil
}
