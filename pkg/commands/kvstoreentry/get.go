package kvstoreentry

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v11/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to fetch the value of a key from an kv store.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.GetKVStoreKeyInput
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get", "Get the value associated with a key")

	// Required.
	c.CmdClause.Flag("key", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.StoreID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	value, err := c.Globals.APIClient.GetKVStoreKey(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		// We are encoding the value of the key here to ensure safe
		// output for binary content along with other outputs.
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		text.Output(out, `{"%s": "%s"}`, c.Input.Key, encodedValue)
		return nil
	}

	if c.Globals.Flags.Verbose {
		// We are encoding the value of the key here to ensure safe
		// output for binary content along with other outputs.
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		text.PrintKVStoreKeyValue(out, "", c.Input.Key, encodedValue)
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	fmt.Fprint(out, value)
	return nil
}
