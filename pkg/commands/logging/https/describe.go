package https

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

// DescribeCommand calls the Fastly API to describe an HTTPS logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetHTTPSInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about an HTTPS logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the HTTPS logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	https, err := c.Globals.APIClient.GetHTTPS(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(https)
		if err != nil {
			return err
		}
		out.Write(data)
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", https.ServiceID)
	}
	fmt.Fprintf(out, "Version: %d\n", https.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", https.Name)
	fmt.Fprintf(out, "URL: %s\n", https.URL)
	fmt.Fprintf(out, "Content type: %s\n", https.ContentType)
	fmt.Fprintf(out, "Header name: %s\n", https.HeaderName)
	fmt.Fprintf(out, "Header value: %s\n", https.HeaderValue)
	fmt.Fprintf(out, "Method: %s\n", https.Method)
	fmt.Fprintf(out, "JSON format: %s\n", https.JSONFormat)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", https.TLSCACert)
	fmt.Fprintf(out, "TLS client certificate: %s\n", https.TLSClientCert)
	fmt.Fprintf(out, "TLS client key: %s\n", https.TLSClientKey)
	fmt.Fprintf(out, "TLS hostname: %s\n", https.TLSHostname)
	fmt.Fprintf(out, "Request max entries: %d\n", https.RequestMaxEntries)
	fmt.Fprintf(out, "Request max bytes: %d\n", https.RequestMaxBytes)
	fmt.Fprintf(out, "Message type: %s\n", https.MessageType)
	fmt.Fprintf(out, "Format: %s\n", https.Format)
	fmt.Fprintf(out, "Format version: %d\n", https.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", https.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", https.Placement)

	return nil
}
