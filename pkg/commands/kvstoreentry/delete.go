package kvstoreentry

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteKeysPoolSize is the goroutine/thread-pool size.
// Each pool will take a 'key' from a channel and issue a DELETE request.
const DeleteKeysPoolSize int = 100

// DeleteKeysMaxErrors is the maximum number of errors we'll allow before
// stopping the goroutines from executing.
const DeleteKeysMaxErrors int = 100

// DeleteCommand calls the Fastly API to delete an kv store.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput
	key argparser.OptionalString

	// NOTE: Public fields can be set via `kv-store delete`.
	DeleteAll bool
	MaxErrors int
	PoolSize  int
	StoreID   string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a key")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.StoreID)

	// Optional.
	c.CmdClause.Flag("all", "Delete all entries within the store").Short('a').BoolVar(&c.DeleteAll)
	c.CmdClause.Flag("concurrency", "The thread pool size (ignored when set without the --all flag)").Default(strconv.Itoa(DeleteKeysPoolSize)).Short('r').IntVar(&c.PoolSize)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("key", "Key name").Short('k').Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("max-errors", "The number of errors to accept before stopping (ignored when set without the --all flag)").Default(strconv.Itoa(DeleteKeysMaxErrors)).Short('m').IntVar(&c.MaxErrors)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// TODO: Support --json for bulk deletions.
	if c.DeleteAll && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidDeleteAllJSONKeyCombo
	}
	if c.DeleteAll && c.key.WasSet {
		return fsterr.ErrInvalidDeleteAllKeyCombo
	}
	if !c.DeleteAll && !c.key.WasSet {
		return fsterr.ErrMissingDeleteAllKeyCombo
	}

	if c.DeleteAll {
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
		return c.DeleteAllKeys(out)
	}

	input := fastly.DeleteKVStoreKeyInput{
		StoreID: c.StoreID,
		Key:     c.key.Value,
	}

	err := c.Globals.APIClient.DeleteKVStoreKey(&input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			Key     string `json:"key"`
			ID      string `json:"store_id"`
			Deleted bool   `json:"deleted"`
		}{
			c.key.Value,
			c.StoreID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted key '%s' from KV Store '%s'", c.key.Value, c.StoreID)
	return nil
}

// DeleteAllKeys deletes all keys within the specified KV Store.
// NOTE: It's a public method as it can be called via `kv-store delete --all`.
func (c *DeleteCommand) DeleteAllKeys(out io.Writer) error {
	spinnerMessage := "Deleting keys"
	var spinner text.Spinner

	var err error
	spinner, err = text.NewSpinner(out)
	if err != nil {
		return err
	}
	err = spinner.Start()
	if err != nil {
		return err
	}
	spinner.Message(spinnerMessage + "...")

	p := c.Globals.APIClient.NewListKVStoreKeysPaginator(&fastly.ListKVStoreKeysInput{
		StoreID: c.StoreID,
	})

	errorsCh := make(chan string, c.MaxErrors)
	keysCh := make(chan string, 1000) // number correlates to pagination page size

	var (
		deleteCount atomic.Uint64
		failedKeys  []string
		wg          sync.WaitGroup
	)

	// We have two separate execution flows happening at once:
	//
	// 1. Pushing keys from pagination data into a key channel.
	// 2. Pulling keys from key channel and issuing API DELETE call.
	//
	// We have a limit on the number of errors. Once that limit is reached we'll
	// stop the 2. set of goroutines.

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(keysCh)
		for p.Next() {
			for _, key := range p.Keys() {
				keysCh <- key
			}
		}
	}()

	for range c.PoolSize {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key := range keysCh {
				err := c.Globals.APIClient.DeleteKVStoreKey(&fastly.DeleteKVStoreKeyInput{StoreID: c.StoreID, Key: key})
				if err != nil {
					select {
					case errorsCh <- key:
					default:
						return // channel is blocked
					}
				}
				spinner.Message(spinnerMessage + "..." + strconv.FormatUint(deleteCount.Add(1), 10))
			}
		}()
	}

	wg.Wait()

	close(errorsCh)
	for err := range errorsCh {
		failedKeys = append(failedKeys, err)
	}

	spinnerMessage = "Deleted keys: " + strconv.FormatUint(deleteCount.Load(), 10)

	if len(failedKeys) > 0 {
		spinner.StopFailMessage(spinnerMessage)
		err := spinner.StopFail()
		if err != nil {
			return fmt.Errorf("failed to stop spinner: %w", err)
		}
		return fmt.Errorf("failed to delete %d keys", len(failedKeys))
	}

	spinner.StopMessage(spinnerMessage)
	if err := spinner.Stop(); err != nil {
		return fmt.Errorf("failed to stop spinner: %w", err)
	}

	text.Success(out, "\nDeleted all keys from KV Store '%s'", c.StoreID)
	return nil
}
