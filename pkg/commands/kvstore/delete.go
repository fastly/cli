package kvstore

import (
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/kvstoreentry"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an kv store.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	deleteAll bool
	maxErrors int
	poolSize  int
	Input     fastly.DeleteKVStoreInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a KV Store")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.StoreID)

	// Optional.
	c.CmdClause.Flag("all", "Delete all entries within the store").Short('a').BoolVar(&c.deleteAll)
	c.CmdClause.Flag("concurrency", "The thread pool size (ignored when set without the --all flag)").Default(strconv.Itoa(kvstoreentry.DeleteKeysPoolSize)).Short('r').IntVar(&c.poolSize)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("max-errors", "The number of errors to accept before stopping (ignored when set without the --all flag)").Default(strconv.Itoa(kvstoreentry.DeleteKeysMaxErrors)).Short('m').IntVar(&c.maxErrors)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.deleteAll {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			text.Warning(out, "This will delete ALL entries from your store!\n\n")
			cont, err := text.AskYesNo(out, "Are you sure you want to continue? [y/N]: ", in)
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}
			text.Break(out)
		}
		dc := kvstoreentry.DeleteCommand{
			Base: argparser.Base{
				Globals: c.Globals,
			},
			DeleteAll: c.deleteAll,
			MaxErrors: c.maxErrors,
			PoolSize:  c.poolSize,
			StoreID:   c.Input.StoreID,
		}
		if err := dc.DeleteAllKeys(out); err != nil {
			return err
		}
		text.Break(out)
	}

	err := c.Globals.APIClient.DeleteKVStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to delete KV store: %w", err)
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.Input.StoreID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted KV Store '%s'", c.Input.StoreID)
	return nil
}
