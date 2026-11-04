package configstoreentry

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// deleteKeysConcurrencyLimit is used to limit the concurrency when deleting ALL keys.
// This is effectively the 'thread pool' size.
const deleteKeysConcurrencyLimit int = 100

// batchLimit is used to split the list of items into batches.
// The batch size of 100 aligns with the KV Store pagination default limit.
const batchLimit int = 100

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	c.input.StoreID = c.StoreID
	c.input.Key = c.Key
	return nil
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	DeleteAll   bool                  `name:"all" short:"a" help:"Delete all entries within the store"`
	BatchSize   argparser.OptionalInt `short:"b" help:"Key batch processing size (ignored when set without the --all flag)"`
	Concurrency argparser.OptionalInt `short:"c" help:"Control thread pool size (ignored when set without the --all flag)"`
	argparser.JSONOutput
	input fastly.DeleteConfigStoreItemInput

	Key string `short:"k" help:"Item name"`
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
	if c.DeleteAll && c.input.Key != "" {
		return fsterr.ErrInvalidDeleteAllKeyCombo
	}
	if !c.DeleteAll && c.input.Key == "" {
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
		return c.deleteAllKeys(out)
	}

	err := c.Globals.APIClient.DeleteConfigStoreItem(context.TODO(), &c.input)
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
	items, err := c.Globals.APIClient.ListConfigStoreItems(context.TODO(), &fastly.ListConfigStoreItemsInput{
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
	if c.Concurrency.WasSet {
		poolSize = c.Concurrency.Value
	}
	semaphore := make(chan struct{}, poolSize)

	total := len(items)
	failedKeys := []string{}

	batchSize := batchLimit
	if c.BatchSize.WasSet {
		batchSize = c.BatchSize.Value
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
				err := c.Globals.APIClient.DeleteConfigStoreItem(context.TODO(), &fastly.DeleteConfigStoreItemInput{StoreID: c.input.StoreID, Key: item.Key})
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
