package kvstoreentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v11/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from an kv store.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.GetKVStoreItemInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Get the associated attributes of a key")

	// Required.
	c.CmdClause.Flag("key", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.StoreID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	item, err := c.Globals.APIClient.GetKVStoreItem(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := map[string]interface{}{
			"key":        c.Input.Key,
			"generation": fmt.Sprintf("%d", item.Generation),
			"metadata":   item.Metadata,
		}
		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}
		return nil
	}

	if c.Globals.Flags.Verbose {
		// Print the key attributes.
		fmt.Fprintf(out, "Key: %s\n", c.Input.Key)
		fmt.Fprintf(out, "Generation: %d\n", item.Generation)
		fmt.Fprintf(out, "Metadata: %s\n", item.Metadata)
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	// Print the key attributes.
	fmt.Fprintf(out, "Key: %s\n", c.Input.Key)
	fmt.Fprintf(out, "Generation: %d\n", item.Generation)
	fmt.Fprintf(out, "Metadata: %s\n", item.Metadata)
	return nil
}
