package kvstoreentry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/runtime"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Insert a key-value pair").Alias("insert")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.StoreID)

	// Optional.
	c.CmdClause.Flag("add", "Limit the operation to adding a new item. If an existing item with the specified key exists, the operation will fail (default: false)").BoolVar(&c.add)
	c.CmdClause.Flag("append", "If an item with the specified key exists, the value provided in the operation is appended to the existing value instead of replacing it (default: false)").BoolVar(&c.append)
	c.CmdClause.Flag("background-fetch", "If set to true, the new value for the item will not be immediately visible to other users of the KV store; they will receive the existing (stale) value while the platform updates cached copies. Setting this to true ensures that other users of the KV store will receive responses to 'get' operations for this item quickly, although they will be slightly out of date (default: false)").BoolVar(&c.backFetch)
	c.CmdClause.Flag("dir", "Path to a directory containing individual files where the filename is the key and the file contents is the value").StringVar(&c.dirPath)
	c.CmdClause.Flag("dir-allow-hidden", "Allow hidden files (e.g. dot files) to be included (skipped by default)").BoolVar(&c.dirAllowHidden)
	c.CmdClause.Flag("dir-concurrency", "Limit the number of concurrent network resources allocated").Default("50").IntVar(&c.dirConcurrency)
	c.CmdClause.Flag("file", `Path to a file containing individual JSON objects (e.g., {"key":"...","value":"base64_encoded_value"}) separated by new-line delimiter`).StringVar(&c.filePath)
	c.CmdClause.Flag("if-generation-match", "Value which must match the current generation marker in an item for an update operation to proceed").StringVar(&c.ifGenMatch)
	c.CmdClause.Flag("metadata", "An arbitrary data field which can contain up to 2000 bytes of data").StringVar(&c.metadata)
	c.CmdClause.Flag("prepend", "If an item with the specified key exists, the value provided in the operation is prepended to the existing value instead of replacing it (Default: false)").BoolVar(&c.prepend)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("key", "Key name").Short('k').StringVar(&c.Input.Key)
	c.CmdClause.Flag("stdin", "Read new-line separated JSON stream via STDIN").BoolVar(&c.stdin)
	c.CmdClause.Flag("value", "Value").StringVar(&c.Input.Value)

	return &c
}

// CreateCommand calls the Fastly API to insert a key into an kv store.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	add            bool
	append         bool
	backFetch      bool
	dirAllowHidden bool
	dirConcurrency int
	dirPath        string
	filePath       string
	ifGenMatch     string
	metadata       string
	prepend        bool
	stdin          bool

	Input fastly.InsertKVStoreKeyInput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if err := c.CheckFlags(); err != nil {
		return err
	}

	if c.stdin {
		return c.ProcessStdin(in, out)
	}

	if c.filePath != "" {
		return c.ProcessFile(out)
	}

	if c.dirPath != "" {
		return c.ProcessDir(in, out)
	}

	if c.Input.Key == "" || c.Input.Value == "" {
		return fsterr.ErrInvalidKVCombo
	}

	// Append optional params.
	c.Input.Add = c.add
	c.Input.Append = c.append
	c.Input.BackgroundFetch = c.backFetch
	// Parse generation match if provided.
	if c.ifGenMatch != "" {
		inputGeneration, err := strconv.ParseUint(c.ifGenMatch, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid generation value: %s", c.ifGenMatch)
		}
		c.Input.IfGenerationMatch = inputGeneration
	}

	if c.metadata != "" {
		c.Input.Metadata = &c.metadata
	}
	c.Input.Prepend = c.prepend

	err := c.Globals.APIClient.InsertKVStoreKey(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}{
			c.Input.StoreID,
			c.Input.Key,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Created key '%s' in KV Store '%s'", c.Input.Key, c.Input.StoreID)
	return nil
}

// CheckFlags ensures only one of the three specified flags are provided.
func (c *CreateCommand) CheckFlags() error {
	flagCount := 0
	if c.stdin {
		flagCount++
	}
	if c.filePath != "" {
		flagCount++
	}
	if c.dirPath != "" {
		flagCount++
	}
	if flagCount > 1 {
		return fsterr.ErrInvalidStdinFileDirCombo
	}
	return nil
}

// ProcessStdin streams STDIN to the batch API endpoint.
func (c *CreateCommand) ProcessStdin(in io.Reader, out io.Writer) error {
	// Determine if 'in' has data available.
	if in == nil || text.IsTTY(in) {
		return fsterr.ErrNoSTDINData
	}
	if c.Globals.Verbose() {
		in = io.TeeReader(in, out)
	}
	return c.CallBatchEndpoint(in, out)
}

// ProcessFile streams a JSON file content to the batch API endpoint.
func (c *CreateCommand) ProcessFile(out io.Writer) error {
	f, err := os.Open(c.filePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	var in io.Reader = f
	if c.Globals.Verbose() {
		in = io.TeeReader(f, out)
	}
	return c.CallBatchEndpoint(in, out)
}

// ProcessDir concurrently reads files from the given directory structure and
// uploads each file to the set-value-for-key endpoint where the filename is the
// key and the file content is the value.
//
// NOTE: Unlike ProcessStdin/ProcessFile content doesn't need to be base64.
func (c *CreateCommand) ProcessDir(in io.Reader, out io.Writer) error {
	if runtime.Windows {
		cont, err := c.PromptWindowsUser(in, out)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		if !cont {
			return nil
		}
		text.Break(out)
	}

	path, err := filepath.Abs(c.dirPath)
	if err != nil {
		return err
	}

	allFiles, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	filteredFiles := make([]fs.DirEntry, 0)
	for _, file := range allFiles {
		// Skip directories/symlinks OR any hidden files unless the user allows them.
		if !file.Type().IsRegular() || (file.Type().IsRegular() && isHiddenFile(file.Name()) && !c.dirAllowHidden) {
			continue
		}
		filteredFiles = append(filteredFiles, file)
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	filesTotal := len(filteredFiles)
	msg := "%s %d of %d files"
	spinner.Message(fmt.Sprintf(msg, "Processing", 0, filesTotal) + "...")

	processed := make(chan struct{}, c.dirConcurrency)
	sem := make(chan struct{}, c.dirConcurrency)
	filesVerboseOutput := make(chan string, filesTotal)

	var (
		processingErrors []ProcessErr
		filesProcessed   uint64
		// NOTE: mu protects access to the 'processingErrors' shared resource.
		// We create multiple goroutines (one for each file) and each one has the
		// potential to mutate the slice by appending new errors to it.
		mu sync.Mutex
		wg sync.WaitGroup
	)

	go func() {
		for range processed {
			atomic.AddUint64(&filesProcessed, 1)
			spinner.Message(fmt.Sprintf(msg, "Processing", filesProcessed, filesTotal) + "...")
		}
	}()

	for _, file := range filteredFiles {
		wg.Add(1)

		go func(file fs.DirEntry) {
			// Restrict resource allocation if concurrency limit is exceeded.
			sem <- struct{}{}
			defer func() {
				processed <- struct{}{}
				<-sem
			}()
			defer wg.Done()

			filename := file.Name()
			filePath := filepath.Join(path, filename)

			if c.Globals.Verbose() {
				filesVerboseOutput <- filename
			}

			// G304 (CWE-22): Potential file inclusion via variable
			// #nosec
			f, err := os.Open(filePath)
			if err != nil {
				mu.Lock()
				processingErrors = append(processingErrors, ProcessErr{
					File: filePath,
					Err:  err,
				})
				mu.Unlock()
				return
			}

			lr, err := fastly.FileLengthReader(f)
			if err != nil {
				mu.Lock()
				processingErrors = append(processingErrors, ProcessErr{
					File: filePath,
					Err:  err,
				})
				mu.Unlock()
				return
			}

			opts := insertKeyOptions{
				client: c.Globals.APIClient,
				id:     c.Input.StoreID,
				key:    filename,
				file:   lr,
			}

			err = insertKey(opts)
			if err != nil {
				// In case the network connection is lost due to exhaustion of
				// resources, then try one more time to make the request.
				//
				// NOTE: you can't type assert the error as it's not exported.
				// https://github.com/golang/go/issues/54173
				if strings.Contains(err.Error(), "net/http: cannot rewind body after connection loss") {
					err = insertKey(opts)
					if err == nil {
						return
					}
				}
				mu.Lock()
				processingErrors = append(processingErrors, ProcessErr{
					File: filePath,
					Err:  err,
				})
				mu.Unlock()
				return
			}
		}(file)
	}

	wg.Wait()

	spinner.StopMessage(fmt.Sprintf(msg, "Processed", atomic.LoadUint64(&filesProcessed)-uint64(len(processingErrors)), filesTotal))
	err = spinner.Stop()
	if err != nil {
		return err
	}

	if c.Globals.Verbose() {
		close(filesVerboseOutput)
		text.Break(out)
		for filename := range filesVerboseOutput {
			fmt.Println(filename)
		}
	}

	if len(processingErrors) == 0 {
		text.Success(out, "\nInserted %d keys into KV Store", len(filteredFiles))
		return nil
	}

	text.Break(out)
	for _, err := range processingErrors {
		fmt.Printf("File: %s\nError: %s\n\n", err.File, err.Err.Error())
	}

	return errors.New("failed to process all the provided files (see error log above ⬆️)")
}

// PromptWindowsUser ensures a user understands that we only filter files whose
// name is prefixed with a dot and not any other kind of 'hidden' attribute that
// can be set by the Windows platform.
func (c *CreateCommand) PromptWindowsUser(in io.Reader, out io.Writer) (bool, error) {
	if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		label := `The Fastly CLI will skip dotfiles (filenames prefixed with a period character, example: '.ignore') but this does not include files set with a "hidden" attribute). Are you sure you want to continue? [y/N] `
		result, err := text.AskYesNo(out, label, in)
		if err != nil {
			return false, err
		}
		return result, nil
	}
	return true, nil
}

// CallBatchEndpoint calls the batch API endpoint.
func (c *CreateCommand) CallBatchEndpoint(in io.Reader, out io.Writer) error {
	type result struct {
		Success bool                  `json:"success"`
		Errors  []*fastly.ErrorObject `json:"errors,omitempty"`
	}

	if err := c.Globals.APIClient.BatchModifyKVStoreKey(context.TODO(), &fastly.BatchModifyKVStoreKeyInput{
		StoreID: c.Input.StoreID,
		Body:    in,
	}); err != nil {
		c.Globals.ErrLog.Add(err)

		r := result{Success: false}

		he, ok := err.(*fastly.HTTPError)
		if ok {
			r.Errors = append(r.Errors, he.Errors...)
		}

		if c.JSONOutput.Enabled {
			_, err := c.WriteJSON(out, r)
			return err
		}

		// If we were able to convert the error into a fastly.HTTPError, then
		// display those errors to the user, otherwise we'll display the original
		// error type.
		if ok {
			for i, e := range he.Errors {
				text.Output(out, "Error %d", i)
				text.Output(out, "Title: %s", e.Title)
				text.Output(out, "Code: %s", e.Code)
				text.Output(out, "Detail: %s", e.Detail)
				text.Break(out)
			}
			return he
		}
		return err
	}

	if c.JSONOutput.Enabled {
		_, err := c.WriteJSON(out, result{Success: true})
		return err
	}

	if c.Globals.Verbose() {
		text.Break(out)
	}
	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func insertKey(opts insertKeyOptions) error {
	return opts.client.InsertKVStoreKey(context.TODO(), &fastly.InsertKVStoreKeyInput{
		Body:    opts.file,
		StoreID: opts.id,
		Key:     opts.key,
	})
}

type insertKeyOptions struct {
	client api.Interface
	id     string
	key    string
	file   fastly.LengthReader
}

// ProcessErr represents an error related to processing individual files.
type ProcessErr struct {
	File string
	Err  error
}
