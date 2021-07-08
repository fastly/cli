package service

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// SearchCommand calls the Fastly API to describe a service.
type SearchCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.SearchServiceInput
}

// NewSearchCommand returns a usable command registered under the parent.
func NewSearchCommand(parent cmd.Registerer, globals *config.Data) *SearchCommand {
	var c SearchCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("search", "Search for a Fastly service by name")
	c.CmdClause.Flag("name", "Service name").Short('n').StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SearchCommand) Exec(in io.Reader, out io.Writer) error {
	service, err := c.Globals.Client.SearchService(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.PrintService(out, "", service)
	return nil
}
