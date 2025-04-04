package kvstore

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an kv store.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.CreateKVStoreInput
}

// locations is a list of supported regional location options.
var locations = []string{"US", "EU", "ASIA", "AUS"}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create a KV Store")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("location", "Regional location of KV Store").Short('l').HintOptions(locations...).EnumVar(&c.Input.Location, locations...)
	c.CmdClause.Flag("name", "Name of KV Store").Short('n').Required().StringVar(&c.Input.Name)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.CreateKVStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created KV Store '%s' (%s)", o.Name, o.StoreID)
	return nil
}
