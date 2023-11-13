package kvstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from an kv store.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input fastly.GetKVStoreInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Describe an kv store").Alias("get")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetKVStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintKVStore(out, "", o)
	return nil
}
