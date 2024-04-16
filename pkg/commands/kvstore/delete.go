package kvstore

import (
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v9/fastly"

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

	batchSize    int
	deleteAll    bool
	poolSize     int
	requestLimit int
	Input        fastly.DeleteKVStoreInput
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
	c.CmdClause.Flag("batch-size", "Splits each thread pool's work into nested concurrent batches (ignored when set without the --all flag)").Default(strconv.Itoa(kvstoreentry.DeleteKeysBatchSize)).Short('b').IntVar(&c.batchSize)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("pool-size", "The thread pool size, each thread handles a maximum of 1000 keys (ignored when set without the --all flag)").Default(strconv.Itoa(kvstoreentry.DeleteKeysPoolSize)).Short('c').IntVar(&c.poolSize)
	c.CmdClause.Flag("request-limit", "The maximum number of API requests to allow (ignored when set without the --all flag)").Default(strconv.Itoa(kvstoreentry.DeleteKeysRequestLimit)).Short('r').IntVar(&c.requestLimit)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.deleteAll {
		dc := kvstoreentry.DeleteCommand{
			Base: argparser.Base{
				Globals: c.Globals,
			},
			BatchSize:    c.batchSize,
			PoolSize:     c.poolSize,
			DeleteAll:    c.deleteAll,
			RequestLimit: c.requestLimit,
			StoreID:      c.Input.StoreID,
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
