package configstoreentry

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// deleteKeysConcurrencyLimit is used to limit the concurrency when deleting ALL keys.
// This is effectively the 'thread pool' size.
const deleteKeysConcurrencyLimit int = 100

// batchLimit is used to split the list of items into batches.
// The batch size of 100 aligns with the KV Store pagination default limit.
const batchLimit int = 100

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("delete", "Delete a config store item")

	// Required.
	c.RegisterFlag(argparser.StoreIDFlag(&c.input.StoreID)) // --store-id

	// Optional.
	c.CmdClause.Flag("all", "Delete all entries within the store").Short('a').BoolVar(&c.deleteAll)
	c.CmdClause.Flag("batch-size", "Key batch processing size (ignored when set without the --all flag)").Short('b').Action(c.batchSize.Set).IntVar(&c.batchSize.Value)
	c.CmdClause.Flag("concurrency", "Control thread pool size (ignored when set without the --all flag)").Short('c').Action(c.concurrency.Set).IntVar(&c.concurrency.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "key",
		Short:       'k',
		Description: "Item name",
		Dst:         &c.input.Key,
	})

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	batchSize   argparser.OptionalInt
	concurrency argparser.OptionalInt
	deleteAll   bool
	input       fastly.DeleteConfigStoreItemInput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// TODO: Support --json for bulk deletions.
	if c.deleteAll && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidDeleteAllJSONKeyCombo
	}
	if c.deleteAll && c.input.Key != "" {
		return fsterr.ErrInvalidDeleteAllKeyCombo
	}
	if !c.deleteAll && c.input.Key == "" {
		return fsterr.ErrMissingDeleteAllKeyCombo
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
		return c.deleteAllKeys(out)
	}

	err := c.Globals.APIClient.DeleteConfigStoreItem(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			StoreID string `json:"store_id"`
			Key     string `json:"key"`
			Deleted bool   `json:"deleted"`
		}{
			c.input.StoreID,
			c.input.Key,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted key '%s' from Config Store '%s'", c.input.Key, c.input.StoreID)
	return nil
}

func (c *DeleteCommand) deleteAllKeys(out io.Writer) error {
	// NOTE: The Config Store returns ALL items (there is no pagination).
	items, err := c.Globals.APIClient.ListConfigStoreItems(&fastly.ListConfigStoreItemsInput{
		StoreID: c.input.StoreID,
	})
	if err != nil {
		return fmt.Errorf("failed to acquire list of Config Store items: %w", err)
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	poolSize := deleteKeysConcurrencyLimit
	if c.concurrency.WasSet {
		poolSize = c.concurrency.Value
	}
	semaphore := make(chan struct{}, poolSize)

	total := len(items)
	failedKeys := []string{}

	batchSize := batchLimit
	if c.batchSize.WasSet {
		batchSize = c.batchSize.Value
	}

	// With KV Store we have pagination support and so that natively provides us a
	// predefined 'batch' size. Because we don't have pagination with the Config
	// Store it means we'll define our own batch size which the user can override.
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		seg := items[i:end]

		wg.Add(1)
		go func(items []*fastly.ConfigStoreItem) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			defer wg.Done()

			for _, item := range items {
				text.Output(out, "Deleting key: %s", item.Key)
				err := c.Globals.APIClient.DeleteConfigStoreItem(&fastly.DeleteConfigStoreItemInput{StoreID: c.input.StoreID, Key: item.Key})
				if err != nil {
					c.Globals.ErrLog.Add(fmt.Errorf("failed to delete key '%s': %s", item.Key, err))
					mu.Lock()
					failedKeys = append(failedKeys, item.Key)
					mu.Unlock()
				}
			}
		}(seg)
	}

	wg.Wait()
	close(semaphore)

	if len(failedKeys) > 0 {
		return fmt.Errorf("failed to delete keys: %s", strings.Join(failedKeys, ", "))
	}

	text.Success(out, "\nDeleted all keys from Config Store '%s'", c.input.StoreID)
	return nil
}
