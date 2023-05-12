package kvstoreentry

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from an kv store.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest manifest.Data
	Input    fastly.GetKVStoreKeyInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Get the value associated with a key").Alias("get")

	// Required.
	c.CmdClause.Flag("key", "Key name").Short('k').Required().StringVar(&c.Input.Key)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)

	// optional
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	value, err := c.Globals.APIClient.GetKVStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		text.Output(out, `{"%s": "%s"}`, c.Input.Key, value)
		return nil
	}

	if c.Globals.Flags.Verbose {
		text.PrintKVStoreKeyValue(out, "", c.Input.Key, value)
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	fmt.Fprint(out, value)
	return nil
}
