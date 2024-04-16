package kvstoreentry

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteKeysPoolSize is the goroutine/thread-pool size when deleting ALL keys.
const DeleteKeysPoolSize int = 1

// DeleteKeysBatchSize is used to split the list of 1000 keys into batches.
// NOTE: The Fastly API returns a maximum of 1000 results per page.
const DeleteKeysBatchSize int = 50

// DeleteKeysRequestLimit is used to restrict the number of open connections.
const DeleteKeysRequestLimit int = 100

// DeleteCommand calls the Fastly API to delete an kv store.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput
	key argparser.OptionalString

	// NOTE: Public fields can be set via `kv-store delete`.
	BatchSize    int
	DeleteAll    bool
	PoolSize     int
	RequestLimit int
	StoreID      string
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
	c.CmdClause.Flag("batch-size", "Splits each thread pool's work into nested concurrent batches (ignored when set without the --all flag)").Default(strconv.Itoa(DeleteKeysBatchSize)).Short('b').IntVar(&c.BatchSize)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("key", "Key name").Short('k').Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("request-limit", "The maximum number of API requests to allow (ignored when set without the --all flag)").Default(strconv.Itoa(DeleteKeysRequestLimit)).Short('c').IntVar(&c.RequestLimit)
	c.CmdClause.Flag("pool-size", "The thread pool size, each thread handles a maximum of 1000 keys (ignored when set without the --all flag)").Default(strconv.Itoa(DeleteKeysPoolSize)).Short('r').IntVar(&c.PoolSize)

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
	p := c.Globals.APIClient.NewListKVStoreKeysPaginator(&fastly.ListKVStoreKeysInput{
		StoreID: c.StoreID,
	})

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	poolSemaphore := make(chan struct{}, c.PoolSize)
	requestSemaphore := make(chan struct{}, c.RequestLimit)
	failedKeysStopCh := make(chan struct{})
	failedKeysCh := make(chan string, 1000)
	failedKeys := []string{}

	spinnerMessage := "deleting keys"
	var spinner text.Spinner

	if !c.Globals.Verbose() {
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
	}

	// Rather than locking a mutex on every single error, we use a goroutine that
	// pulls from the channel whenever it's full.
	go func() {
		for {
			select {
			case <-failedKeysStopCh:
				return
			default:
				if len(failedKeysCh) == cap(failedKeysCh) {
					mu.Lock()
					for range len(failedKeysCh) {
						failedKeys = append(failedKeys, <-failedKeysCh)
					}
					mu.Unlock()
				}
			}
		}
	}()

	// EXAMPLE USAGE:
	//
	// --pool-size 1 --batch-size 50 --request-limit 50
	//
	// EXPLANATION: given 5000 keys.
	//
	// --pool-size 1
	//
	// We restrict the number of threads in the pool to 1.
	// The Fastly API returns 1000 keys per pagination page.
	// This means we only process 1000 keys at a time.
	// The pool size, is how many 'pages' we'll process concurrently.
	//
	// --batch-size 50
	//
	// We divide up the 1000 keys into separate batches of 50.
	// This would mean concurrently processing 1000 keys across 20 threads.
	// So each thread would process 50 API requests across those 20 threads.
	// e.g. 1000/50=20
	//
	// --request-limit 50
	//
	// Although we're allowing 20 threads to handle 50 requests per thread, we are
	// actually restricting the total number of API requests to 50.

	for p.Next() {
		// IMPORTANT: Use copies of the keys when processing data concurrently.
		keys := p.Keys()
		copiedKeys := make([]string, len(keys))
		copy(copiedKeys, keys)

		poolSemaphore <- struct{}{}
		wg.Add(1)

		go func(keys []string) {
			var wgBatch sync.WaitGroup
			defer func() { <-poolSemaphore }()
			defer wg.Done()

			total := len(keys)
			for i := 0; i < total; i += c.BatchSize {
				end := i + c.BatchSize
				if end > total {
					end = total
				}
				seg := keys[i:end]

				wgBatch.Add(1)
				go func(seg []string) {
					defer wgBatch.Done()
					for _, key := range seg {
						if c.Globals.Verbose() {
							text.Output(out, "Deleting key: %s", key)
						}
						requestSemaphore <- struct{}{}
						err := c.Globals.APIClient.DeleteKVStoreKey(&fastly.DeleteKVStoreKeyInput{StoreID: c.StoreID, Key: key})
						<-requestSemaphore
						if err != nil {
							c.Globals.ErrLog.Add(fmt.Errorf("failed to delete key '%s': %s", key, err))
							failedKeysCh <- key
						}
					}
				}(seg)
			}

			// The paginator calls .Next() even when there are zero keys.
			// So to avoid a deadlock we check the key length before waiting.
			if total > 0 {
				wgBatch.Wait()
			}
		}(copiedKeys)
	}

	wg.Wait()
	close(poolSemaphore)
	close(requestSemaphore)
	close(failedKeysCh)
	failedKeysStopCh <- struct{}{}
	if len(failedKeysCh) > 0 {
		for range len(failedKeysCh) {
			failedKeys = append(failedKeys, <-failedKeysCh)
		}
	}

	// The pagination might have an error, and so we'll make sure to print any
	// failed API requests at the same time.
	if err := p.Err(); err != nil {
		if len(failedKeys) > 0 {
			err = fmt.Errorf("failed to delete keys (error: %s) when handling a pagination error: %w)", strings.Join(failedKeys, ", "), err)
		}
		retErr := fmt.Errorf("failed to paginate keys: %w", err)
		if !c.Globals.Verbose() {
			spinner.StopFailMessage(spinnerMessage)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf("failed to stop spinner (error: %w) when handling the error: %w", spinErr, retErr)
			}
			return retErr
		}
	}

	// There could be failed API requests even if the pagination didn't fail
	if len(failedKeys) > 0 {
		retErr := fmt.Errorf("failed to delete keys: %s", strings.Join(failedKeys, ", "))
		if !c.Globals.Verbose() {
			spinner.StopFailMessage(spinnerMessage)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf("failed to stop spinner (error: %w) when handling the error: %w", spinErr, retErr)
			}
			return retErr
		}
		return retErr
	}

	if !c.Globals.Verbose() {
		spinner.StopMessage(spinnerMessage)
		_ = spinner.Stop()
	}

	text.Success(out, "\nDeleted all keys from KV Store '%s'", c.StoreID)
	return nil
}
