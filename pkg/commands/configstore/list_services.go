package configstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListServicesCommand returns a usable command registered under the parent.
func NewListServicesCommand(parent argparser.Registerer, g *global.Data) *ListServicesCommand {
	c := ListServicesCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("list-services", "List config store's services")

	// Required.
	c.RegisterFlag(argparser.StoreIDFlag(&c.input.StoreID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// ListServicesCommand calls the Fastly API to list appropriate resources.
type ListServicesCommand struct {
	argparser.Base
	argparser.JSONOutput
	input fastly.ListConfigStoreServicesInput
}

// Exec invokes the application logic for the command.
func (c *ListServicesCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.ListConfigStoreServices(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintConfigStoreServicesTbl(out, o)

	return nil
}
