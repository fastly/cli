package kvstoreentry

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Insert a key-value pair").Alias("insert")

	// Required.
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)

	// Optional.
	c.CmdClause.Flag("dir", "Path to a directory containing individual files where the filename is the key and the file contents is the value").StringVar(&c.dirPath)
	c.CmdClause.Flag("dir-allow-hidden", "Allow hidden files (e.g. dot files) to be included (skipped by default)").BoolVar(&c.dirAllowHidden)
	c.CmdClause.Flag("dir-concurrency", "Limit the number of concurrent network resources allocated").Default("50").IntVar(&c.dirConcurrency)
	c.CmdClause.Flag("file", "Path to a file containing individual JSON objects separated by new-line delimiter").StringVar(&c.filePath)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("key", "Key name").Short('k').StringVar(&c.Input.Key)
	c.CmdClause.Flag("stdin", "Read new-line separated JSON stream via STDIN").BoolVar(&c.stdin)
	c.CmdClause.Flag("value", "Value").StringVar(&c.Input.Value)

	return &c
}

// CreateCommand calls the Fastly API to insert a key into an kv store.
type CreateCommand struct {
	cmd.Base
	cmd.JSONOutput

	dirAllowHidden bool
	dirConcurrency int
	dirPath        string
	filePath       string
	manifest       manifest.Data
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
		return c.ProcessDir(out)
	}

	if c.Input.Key == "" || c.Input.Value == "" {
		return fsterr.ErrInvalidKVCombo
	}

	err := c.Globals.APIClient.InsertKVStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}{
			c.Input.ID,
			c.Input.Key,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Inserted key %s into KV Store %s", c.Input.Key, c.Input.ID)

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
	return c.CallBatchEndpoint(f, out)
}

// ProcessDir concurrently reads files from the given directory structure and
// uploads each file to the set-value-for-key endpoint where the filename is the
// key and the file content is the value.
//
// NOTE: Unlike ProcessStdin/ProcessFile content doesn't need to be base64.
func (c *CreateCommand) ProcessDir(out io.Writer) error {
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
		hidden, err := isHiddenFile(file.Name())
		if err != nil {
			return err
		}
		// Skip directories/symlinks OR any hidden files unless the user allows them.
		if !file.Type().IsRegular() || (file.Type().IsRegular() && hidden && !c.dirAllowHidden) {
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
	fileLength := len(filteredFiles)
	msg := "%s %d of %d files"
	spinner.Message(fmt.Sprintf(msg, "Processing", 0, fileLength) + "...")

	base := filepath.Base(path)
	processed := make(chan struct{}, c.dirConcurrency)
	sem := make(chan struct{}, c.dirConcurrency)
	done := make(chan bool)

	var (
		processingErrors []ProcessErr
		filesProcessed   uint64
		mu               sync.Mutex
		wg               sync.WaitGroup
	)

	go func() {
		for range processed {
			filesProcessed++
			spinner.Message(fmt.Sprintf(msg, "Processing", filesProcessed, fileLength) + "...")
			if filesProcessed >= uint64(len(filteredFiles)) {
				done <- true
			}
		}
	}()

	for _, file := range filteredFiles {
		wg.Add(1)

		go func(file fs.DirEntry) {
			// Restrict resource allocation if concurrency limit is exceeded.
			sem <- struct{}{}
			defer func() {
				// IMPORTANT: Always mark a file as processed regardless of errors.
				// This is so that later we can unblock the 'done' channel.
				// Refer to the earlier goroutine that ranges the 'processed' channel.
				processed <- struct{}{}
				<-sem
			}()
			defer wg.Done()

			filePath := filepath.Join(c.dirPath, file.Name())
			dir, filename := filepath.Split(filePath)
			index := strings.Index(dir, base)
			filename = filepath.Join(dir[index:], filename)

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

			fi, err := f.Stat()
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
				id:     c.Input.ID,
				key:    filename,
				file:   f,
				size:   fi.Size(),
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

	// NOTE: Block via chan to allow final goroutine to increment filesProcessed.
	// Otherwise the StopMessage below is called before filesProcessed is updated.
	<-done

	spinner.StopMessage(fmt.Sprintf(msg, "Processed", filesProcessed-uint64(len(processingErrors)), fileLength))
	err = spinner.Stop()
	if err != nil {
		return err
	}

	if len(processingErrors) == 0 {
		text.Success(out, "Inserted %d keys into KV Store", len(filteredFiles))
		return nil
	}

	text.Break(out)
	for _, err := range processingErrors {
		fmt.Printf("File: %s\nError: %s\n\n", err.File, err.Err.Error())
	}

	return errors.New("failed to process all the provided files (see error log above ⬆️)")
}

// CallBatchEndpoint calls the batch API endpoint.
func (c *CreateCommand) CallBatchEndpoint(in io.Reader, out io.Writer) error {
	if err := c.Globals.APIClient.BatchModifyKVStoreKey(&fastly.BatchModifyKVStoreKeyInput{
		ID:   c.Input.ID,
		Body: in,
	}); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func insertKey(opts insertKeyOptions) error {
	return opts.client.InsertKVStoreKey(&fastly.InsertKVStoreKeyInput{
		Body:       opts.file,
		BodyLength: opts.size,
		ID:         opts.id,
		Key:        opts.key,
	})
}

type insertKeyOptions struct {
	client api.Interface
	id     string
	key    string
	file   io.Reader
	size   int64
}

type ProcessErr struct {
	File string
	Err  error
}
