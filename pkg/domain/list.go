package domain

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list domains.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListDomainsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List domains on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	domains, err := c.Globals.Client.ListDomains(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME", "COMMENT")
		for _, domain := range domains {
			tw.AddLine(domain.ServiceID, domain.Version, domain.Name, domain.Comment)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, domain := range domains {
		fmt.Fprintf(out, "\tDomain %d/%d\n", i+1, len(domains))
		fmt.Fprintf(out, "\t\tName: %s\n", domain.Name)
		fmt.Fprintf(out, "\t\tComment: %v\n", domain.Comment)
	}
	fmt.Fprintln(out)

	return nil
}
