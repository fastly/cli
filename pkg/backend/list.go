package backend

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list backends.
type ListCommand struct {
	common.Base
	Input fastly.ListBackendsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List backends on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.Input.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	backends, err := c.Globals.Client.ListBackends(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME", "ADDRESS", "PORT", "COMMENT")
		for _, backend := range backends {
			tw.AddLine(backend.ServiceID, backend.ServiceVersion, backend.Name, backend.Address, backend.Port, backend.Comment)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, backend := range backends {
		fmt.Fprintf(out, "\tBackend %d/%d\n", i+1, len(backends))
		text.PrintBackend(out, "\t\t", backend)
	}
	fmt.Fprintln(out)

	return nil
}
