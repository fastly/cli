package https

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list HTTPS logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListHTTPSInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List HTTPS endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	httpss, err := c.Globals.Client.ListHTTPS(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, https := range httpss {
			tw.AddLine(https.ServiceID, https.ServiceVersion, https.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, https := range httpss {
		fmt.Fprintf(out, "\tHTTPS %d/%d\n", i+1, len(httpss))
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
