package https

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list HTTPS logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.ListHTTPSInput
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
	c.CmdClause = parent.Command("list", "List HTTPS endpoints on a Fastly service version")

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

	o, err := c.Globals.APIClient.ListHTTPS(&c.Input)
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
		for _, https := range o {
			tw.AddLine(https.ServiceID, https.ServiceVersion, https.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, https := range o {
		fmt.Fprintf(out, "\tHTTPS %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", https.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", https.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", https.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", https.URL)
		fmt.Fprintf(out, "\t\tContent type: %s\n", https.ContentType)
		fmt.Fprintf(out, "\t\tHeader name: %s\n", https.HeaderName)
		fmt.Fprintf(out, "\t\tHeader value: %s\n", https.HeaderValue)
		fmt.Fprintf(out, "\t\tMethod: %s\n", https.Method)
		fmt.Fprintf(out, "\t\tJSON format: %s\n", https.JSONFormat)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", https.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS client certificate: %s\n", https.TLSClientCert)
		fmt.Fprintf(out, "\t\tTLS client key: %s\n", https.TLSClientKey)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", https.TLSHostname)
		fmt.Fprintf(out, "\t\tRequest max entries: %d\n", https.RequestMaxEntries)
		fmt.Fprintf(out, "\t\tRequest max bytes: %d\n", https.RequestMaxBytes)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", https.MessageType)
		fmt.Fprintf(out, "\t\tFormat: %s\n", https.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", https.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", https.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", https.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
