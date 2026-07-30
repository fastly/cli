package integration

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListTypesCommand calls the Fastly API to list the supported notification
// integration types and the configuration each one requires.
type ListTypesCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// NewListTypesCommand returns a usable command registered under the parent.
func NewListTypesCommand(parent argparser.Registerer, g *global.Data) *ListTypesCommand {
	c := ListTypesCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list-types", "List supported notification integration types")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListTypesCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetIntegrationTypes(context.TODO())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	var types []fastly.IntegrationType
	if o != nil {
		types = *o
	}
	text.PrintIntegrationTypesTbl(out, types)

	return nil
}
