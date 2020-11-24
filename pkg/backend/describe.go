package backend

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// DescribeCommand calls the Fastly API to describe a backend.
type DescribeCommand struct {
	common.Base
	Input fastly.GetBackendInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.CmdClause = parent.Command("describe", "Show detailed information about a backend on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "Name of backend").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	backend, err := c.Globals.Client.GetBackend(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", backend.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", backend.ServiceVersion)
	text.PrintBackend(out, "", backend)

	return nil
}
