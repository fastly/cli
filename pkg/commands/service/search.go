package service

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// SearchCommand calls the Fastly API to describe a service.
type SearchCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.SearchServiceInput
}

// NewSearchCommand returns a usable command registered under the parent.
func NewSearchCommand(parent argparser.Registerer, g *global.Data) *SearchCommand {
	c := SearchCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("search", "Search for a Fastly service by name")

	// Required.
	c.CmdClause.Flag("name", "Service name").Short('n').Required().StringVar(&c.Input.Name)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *SearchCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	service, err := c.Globals.APIClient.SearchService(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Name": c.Input.Name,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, service); ok {
		return err
	}

	text.PrintService(out, "", service)
	return nil
}
