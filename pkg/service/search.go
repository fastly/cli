package service

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// SearchCommand calls the Fastly API to describe a service.
type SearchCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.SearchServiceInput
}

// NewSearchCommand returns a usable command registered under the parent.
func NewSearchCommand(parent common.Registerer, globals *config.Data) *SearchCommand {
	var c SearchCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("search", "Search for a Fastly service by name")
	c.CmdClause.Flag("name", "Service Name").Short('n').StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SearchCommand) Exec(in io.Reader, out io.Writer) error {
	service, err := c.Globals.Client.SearchService(&c.Input)
	if err != nil {
		return err
	}

	text.PrintService(out, "", service)
	return nil
}
