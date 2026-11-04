package kvstore

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a kv store.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.CreateKVStoreInput `kong:"-"`

	Location string `short:"l" default:"" enum:"US,EU,ASIA,AUS," hints:"US,EU,ASIA,AUS" help:"Regional location of KV Store"`
	NameFlag string `name:"name" short:"n" required:"" help:"Name of KV Store"`
}

// locations is a list of supported regional location options.
var locations = []string{"US", "EU", "ASIA", "AUS"}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.Input.Location = c.Location
	c.Input.Name = c.NameFlag
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.CreateKVStore(context.TODO(), &c.Input)
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
