package kvstoreentry

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// deleteKeysConcurrencyLimit is used to limit the concurrency when deleting ALL keys.
const deleteKeysConcurrencyLimit = 1000

// DeleteCommand calls the Fastly API to delete an kv store.
type DeleteCommand struct {
	cmd.Base
	cmd.JSONOutput

	deleteAll bool
	key       cmd.OptionalString
	manifest  manifest.Data
	storeID   string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("delete", "Delete a key")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.storeID)

	// Optional.
	c.CmdClause.Flag("all", "Delete all entries within the store").Short('a').BoolVar(&c.deleteAll)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("key", "Key name").Short('k').Action(c.key.Set).StringVar(&c.key.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if c.deleteAll && c.key.WasSet {
		return fsterr.ErrInvalidDeleteAllKeyCombo
	}
	if !c.deleteAll && !c.key.WasSet {
		return fsterr.ErrMissingDeleteAllKeyCombo
	}

	if c.deleteAll {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			text.Warning(out, "This will delete ALL entries from your store!")
			text.Break(out)
			cont, err := text.AskYesNo(out, "Are you sure you want to continue? [yes/no]: ", in)
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

	input := fastly.DeleteKVStoreKeyInput{
		ID:  c.storeID,
		Key: c.key.Value,
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
			c.storeID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted key '%s' from KV Store '%s'", c.key.Value, c.storeID)
	return nil
}

func (c *DeleteCommand) deleteAllKeys(out io.Writer) error {
	p := c.Globals.APIClient.NewListKVStoreKeysPaginator(&fastly.ListKVStoreKeysInput{
		ID: c.storeID,
	})

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	semaphore := make(chan struct{}, deleteKeysConcurrencyLimit)

	failedKeys := []string{}

	for p.Next() {
		// IMPORTANT: Use copies of the keys when processing data concurrently.
		keys := p.Keys()
		copiedKeys := make([]string, len(keys))
		copy(copiedKeys, keys)

		wg.Add(1)
		go func(keys []string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			defer wg.Done()

			sort.Strings(keys)
			for _, key := range keys {
				text.Output(out, "Deleting key: %s", key)
				err := c.Globals.APIClient.DeleteKVStoreKey(&fastly.DeleteKVStoreKeyInput{ID: c.storeID, Key: key})
				if err != nil {
					c.Globals.ErrLog.Add(fmt.Errorf("failed to delete key '%s': %s", key, err))
					mu.Lock()
					failedKeys = append(failedKeys, key)
					mu.Unlock()
				}
			}
		}(keys)
	}

	wg.Wait()
	close(semaphore)

	if err := p.Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %s", err)
	}
	if len(failedKeys) > 0 {
		return fmt.Errorf("failed to delete keys: %s", strings.Join(failedKeys, ", "))
	}

	text.Success(out, "Deleted all keys from KV Store '%s'", c.storeID)
	return nil
}
