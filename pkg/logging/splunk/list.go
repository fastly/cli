package splunk

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// ListCommand calls the Fastly API to list Splunk logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListSplunksInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Splunk endpoints on a Fastly service version")
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

	splunks, err := c.Globals.Client.ListSplunks(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, splunk := range splunks {
			tw.AddLine(splunk.ServiceID, splunk.ServiceVersion, splunk.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, splunk := range splunks {
		fmt.Fprintf(out, "\tSplunk %d/%d\n", i+1, len(splunks))
		fmt.Fprintf(out, "\t\tService ID: %s\n", splunk.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", splunk.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", splunk.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", splunk.URL)
		fmt.Fprintf(out, "\t\tToken: %s\n", splunk.Token)
		fmt.Fprintf(out, "\t\tTLS CA certificate: %s\n", splunk.TLSCACert)
		fmt.Fprintf(out, "\t\tTLS hostname: %s\n", splunk.TLSHostname)
		fmt.Fprintf(out, "\t\tFormat: %s\n", splunk.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", splunk.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", splunk.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", splunk.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
