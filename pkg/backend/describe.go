package backend

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a backend.
type DescribeCommand struct {
	cmd.Base
	Input          fastly.GetBackendInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.CmdClause = parent.Command("describe", "Show detailed information about a backend on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.ServiceID)
	c.NewServiceVersionFlag(cmd.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.CmdClause.Flag("name", "Name of backend").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	v, err := c.serviceVersion.Parse(c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.Input.ServiceVersion = v.Number

	backend, err := c.Globals.Client.GetBackend(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", backend.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", backend.ServiceVersion)
	text.PrintBackend(out, "", backend)

	return nil
}
