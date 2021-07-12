package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	cmd.Base
	Input fastly.ListServicesInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List Fastly services")
	// no flags, because ListServicesInput has no fields
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	services, err := c.Globals.Client.ListServices(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("NAME", "ID", "TYPE", "ACTIVE VERSION", "LAST EDITED (UTC)")
		for _, service := range services {
			updatedAt := "n/a"
			if service.UpdatedAt != nil {
				updatedAt = service.UpdatedAt.UTC().Format(time.Format)
			}

			activeVersion := fmt.Sprint(service.ActiveVersion)
			for _, v := range service.Versions {
				if uint(v.Number) == service.ActiveVersion && !v.Active {
					activeVersion = "n/a"
				}
			}

			tw.AddLine(service.Name, service.ID, text.ServiceType(service.Type), activeVersion, updatedAt)
		}
		tw.Print()
		return nil
	}

	for i, service := range services {
		fmt.Fprintf(out, "Service %d/%d\n", i+1, len(services))
		text.PrintService(out, "\t", service)
		fmt.Fprintln(out)
	}

	return nil
}
