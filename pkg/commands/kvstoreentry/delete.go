package kvstoreentry

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteKeysPoolSize is the goroutine/thread-pool size.
// Each pool will take a 'key' from a channel and issue a DELETE request.
const DeleteKeysPoolSize int = 100

// DeleteKeysMaxErrors is the maximum number of errors we'll allow before
// stopping the goroutines from executing.
const DeleteKeysMaxErrors int = 100

// DeleteCommand calls the Fastly API to delete a kv store.
type DeleteCommand struct {
	argparser.Base

	// Required.

	// NOTE: Public fields can be set via `kv-store delete`.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	DeleteAll         bool   `name:"all" short:"a" help:"Delete all entries within the store"`
	PoolSize          int    `name:"concurrency" short:"r" default:"100" help:"The thread pool size (ignored when set without the --all flag)"`
	Force             bool   `help:"Return a successful result from a 'delete' operation even if the specified key was not found"`
	IfGenerationMatch string `help:"Value which must match the current generation marker in an item for a delete operation to proceed"`
	argparser.JSONOutput
	Key       argparser.OptionalString `short:"k" help:"Key name"`
	Prefix    argparser.OptionalString `help:"Delete items whose keys match this prefix"`
	MaxErrors int                      `short:"m" default:"100" help:"The number of errors to accept before stopping (ignored when set without the --all flag)"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// TODO: Support --json for bulk deletions.
	if (c.DeleteAll || c.Prefix.WasSet) && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidDeleteMultipleJSONKeyCombo
	}
	if c.DeleteAll && c.Key.WasSet {
		return fsterr.ErrInvalidDeleteMultipleKeyCombo
	}
	if c.DeleteAll && c.Prefix.WasSet {
		return fsterr.ErrInvalidDeleteMultipleKeyCombo
	}
	if c.Prefix.WasSet && c.Key.WasSet {
		return fsterr.ErrInvalidDeleteMultipleKeyCombo
	}
	if !c.DeleteAll && !c.Prefix.WasSet && !c.Key.WasSet {
		return fsterr.ErrMissingDeleteMultipleKeyCombo
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
		return c.DeleteMultipleKeys(out, "")
	}

	if c.Prefix.WasSet {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			text.Warning(out, "This may delete MANY entries from your store!\n\n")
			cont, err := text.AskYesNo(out, "Are you sure you want to continue? [y/N]: ", in)
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}
			text.Break(out)
		}
		return c.DeleteMultipleKeys(out, c.Prefix.Value)
	}

	input := fastly.DeleteKVStoreKeyInput{
		StoreID: c.StoreID,
		Force:   c.Force,
		Key:     c.Key.Value,
	}

	// Validate generation value if provided.
	if c.IfGenerationMatch != "" {
		_, err := strconv.ParseUint(c.IfGenerationMatch, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid generation value: %s", c.IfGenerationMatch)
		}
		input.IfGenerationMatch, _ = strconv.ParseUint(c.IfGenerationMatch, 10, 64)
	}

	err := c.Globals.APIClient.DeleteKVStoreKey(context.TODO(), &input)
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
			c.Key.Value,
			c.StoreID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted key '%s' from KV Store '%s'", c.Key.Value, c.StoreID)
	return nil
}

// DeleteMultipleKeys deletes multiple keys from the specified KV Store.
// NOTE: It's a public method as it can be called via `kv-store delete --all`.
func (c *DeleteCommand) DeleteMultipleKeys(out io.Writer, prefix string) error {
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

	p := c.Globals.APIClient.NewListKVStoreKeysPaginator(context.TODO(), &fastly.ListKVStoreKeysInput{
		StoreID: c.StoreID,
		Prefix:  prefix,
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
				err := c.Globals.APIClient.DeleteKVStoreKey(context.TODO(), &fastly.DeleteKVStoreKeyInput{StoreID: c.StoreID, Key: key})
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

	if prefix == "" {
		text.Success(out, "\nDeleted all keys from KV Store '%s'", c.StoreID)
	} else {
		text.Success(out, "\nDeleted %s keys from KV Store '%s'", strconv.FormatUint(deleteCount.Load(), 10), c.StoreID)
	}
	return nil
}
