package kvstoreentry

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
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
	c.CmdClause.Flag("dir", "Path to a directory containing individual files where the filename is the key and the file contents is the value").StringVar(&c.dirPath)
	c.CmdClause.Flag("dir-concurrency", "Limit the number of concurrent network resources allocated").Default("100").IntVar(&c.dirConcurrency)
	c.CmdClause.Flag("file", "Path to a file containing individual JSON objects separated by new-line delimiter").StringVar(&c.filePath)
	c.CmdClause.Flag("key-name", "Key name").Short('k').StringVar(&c.Input.Key)
	c.CmdClause.Flag("stdin", "Read new-line separated JSON stream via STDIN").BoolVar(&c.stdin)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("value", "Value").StringVar(&c.Input.Value)
	return &c
}

// CreateCommand calls the Fastly API to insert a key into an kv store.
type CreateCommand struct {
	cmd.Base
	dirConcurrency int
	dirPath        string
	filePath       string
	manifest       manifest.Data
	stdin          bool

	Input fastly.InsertKVStoreKeyInput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
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

func (c *CreateCommand) ProcessStdin(in io.Reader, out io.Writer) error {
	// Determine if 'in' has data available.
	if in == nil || text.IsTTY(in) {
		return fsterr.ErrNoSTDINData
	}

	if err := c.CallBatchEndpoint(in); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func (c *CreateCommand) ProcessFile(out io.Writer) error {
	f, err := os.Open(c.filePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if err := c.CallBatchEndpoint(f); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func (c *CreateCommand) ProcessDir(out io.Writer) error {
	path, err := filepath.Abs(c.dirPath)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	fileLength := len(files)
	msg := "Processed %d of %d files"
	spinner.Message(fmt.Sprintf(msg, 0, fileLength) + "...")

	base := filepath.Base(path)
	errors := []ProcessErr{}
	processed := make(chan struct{}, c.dirConcurrency)
	sem := make(chan struct{}, c.dirConcurrency)

	var (
		filesProcessed uint64
		wg             sync.WaitGroup
	)

	go func() {
		for range processed {
			filesProcessed++
			spinner.Message(fmt.Sprintf(msg, filesProcessed, fileLength) + "...")
		}
	}()

	for _, file := range files {
		if !file.IsDir() {
			wg.Add(1)

			go func(file fs.DirEntry) {
				// Restrict resource allocation if concurrency limit is exceeded.
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				defer wg.Done()

				filePath := filepath.Join(c.dirPath, file.Name())
				dir, filename := filepath.Split(filePath)
				index := strings.Index(dir, base)
				filename = filepath.Join(dir[index:], filename)

				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					errors = append(errors, ProcessErr{
						File: filePath,
						Err:  err,
					})
					return
				}

				err = c.Globals.APIClient.InsertKVStoreKey(&fastly.InsertKVStoreKeyInput{
					ID:    c.Input.ID,
					Key:   filename,
					Value: string(fileContent),
				})
				if err != nil {
					errors = append(errors, ProcessErr{
						File: filePath,
						Err:  err,
					})
					return
				}

				processed <- struct{}{}
			}(file)
		}
	}

	wg.Wait()

	// NOTE: We add a one second sleep to allow the final goroutine to increment.
	// Otherwise the StopMessage below is called before filesProcessed is updated.
	time.Sleep(time.Second * 1)

	spinner.StopMessage(fmt.Sprintf(msg, filesProcessed, fileLength))
	err = spinner.Stop()
	if err != nil {
		return err
	}

	if len(errors) == 0 {
		text.Success(out, "Inserted %d keys into KV Store", len(files))
		return nil
	}

	text.Error(out, "Inserted %d keys into KV Store", len(files)-len(errors))

	for _, err := range errors {
		fmt.Printf("File: %s\nError: %s\n\n", err.File, err.Err.Error())
	}

	return nil
}

func (c *CreateCommand) CallBatchEndpoint(in io.Reader) error {
	host, _ := c.Globals.Endpoint()
	path := fmt.Sprintf("/resources/stores/kv/%s/batch", c.Input.ID)
	token, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return fsterr.ErrNoToken
	}

	// IMPORTANT: Input could be large so we must buffer the reads.
	// This will avoid loading all of the data into memory at once.
	body := bufio.NewReader(in)

	resp, err := undocumented.Call(
		host, path, http.MethodPut, token, body, c.Globals.HTTPClient,
		undocumented.HTTPHeader{
			Key:   "Content-Type",
			Value: "application/x-ndjson",
		},
	)
	if err != nil {
		apiErr, ok := err.(undocumented.APIError)
		if !ok {
			return err
		}
		return fmt.Errorf("%w: %d %s: %s", err, apiErr.StatusCode, http.StatusText(apiErr.StatusCode), string(resp))
	}

	return nil
}

type ProcessErr struct {
	File string
	Err  error
}
