package splunk

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v2/fastly"
)

// DescribeCommand calls the Fastly API to describe a Splunk logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetSplunkInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Splunk logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Splunk logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	splunk, err := c.Globals.Client.GetSplunk(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", splunk.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", splunk.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", splunk.Name)
	fmt.Fprintf(out, "URL: %s\n", splunk.URL)
	fmt.Fprintf(out, "Token: %s\n", splunk.Token)
	fmt.Fprintf(out, "TLS CA certificate: %s\n", splunk.TLSCACert)
	fmt.Fprintf(out, "TLS hostname: %s\n", splunk.TLSHostname)
	fmt.Fprintf(out, "Format: %s\n", splunk.Format)
	fmt.Fprintf(out, "Format version: %d\n", splunk.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", splunk.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", splunk.Placement)

	return nil
}
