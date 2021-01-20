package logentries

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Logentries logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetLogentriesInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Logentries logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "The name of the Logentries logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	logentries, err := c.Globals.Client.GetLogentries(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", logentries.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", logentries.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", logentries.Name)
	fmt.Fprintf(out, "Port: %d\n", logentries.Port)
	fmt.Fprintf(out, "Use TLS: %t\n", logentries.UseTLS)
	fmt.Fprintf(out, "Token: %s\n", logentries.Token)
	fmt.Fprintf(out, "Format: %s\n", logentries.Format)
	fmt.Fprintf(out, "Format version: %d\n", logentries.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", logentries.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", logentries.Placement)

	return nil
}
