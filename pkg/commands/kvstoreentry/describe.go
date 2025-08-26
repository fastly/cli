package kvstoreentry

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fastly/go-fastly/v11/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from an kv store.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input      fastly.GetKVStoreItemInput
	Generation bool
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Get the value associated with a key").Alias("get")

	// Required.
	c.CmdClause.Flag("key", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.StoreID)

	// Optional.
	c.CmdClause.Flag("generation", "Determines whether the generation marker emits").BoolVar(&c.Generation)
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

	value, err := item.ValueAsString()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		if c.Generation {
			text.Output(out, `{"%s": "%s", "generation": %d}`, c.Input.Key, value, item.Generation)
		} else {
			text.Output(out, `{"%s": "%s"}`, c.Input.Key, value)
		}
		return nil
	}

	if c.Globals.Flags.Verbose {
		text.PrintKVStoreKeyValue(out, "", c.Input.Key, value)
		if c.Generation {
			fmt.Fprintf(out, "Generation: %d\n", item.Generation)
		}
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	fmt.Fprint(out, value)
	// Print the Generation ID if the flag is present.
	if c.Generation {
		fmt.Fprintf(os.Stderr, "\nGeneration: %d\n", item.Generation)
	}
	return nil
}
