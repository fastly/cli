package kvstoreentry

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fastly/go-fastly/v8/fastly"

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
	processed := make(chan struct{}, c.dirConcurrency)
	sem := make(chan struct{}, c.dirConcurrency)
	done := make(chan bool)

	var (
		errors         []ProcessErr
		filesProcessed uint64
		wg             sync.WaitGroup
	)

	go func() {
		for range processed {
			filesProcessed++
			spinner.Message(fmt.Sprintf(msg, filesProcessed, fileLength) + "...")
			if filesProcessed >= uint64(len(files)) {
				done <- true
			}
		}
	}()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

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

			// G304 (CWE-22): Potential file inclusion via variable
			// #nosec
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

	wg.Wait()

	// NOTE: Block via chan to allow final goroutine to increment filesProcessed.
	// Otherwise the StopMessage below is called before filesProcessed is updated.
	<-done

	spinner.StopMessage(fmt.Sprintf(msg, filesProcessed, fileLength))
	err = spinner.Stop()
	if err != nil {
		return err
	}

	outputMsg := "Inserted %d keys into KV Store"

	if len(errors) == 0 {
		text.Success(out, outputMsg, len(files))
		return nil
	}

	text.Error(out, "Encountered the following errors")
	text.Break(out)
	for _, err := range errors {
		fmt.Printf("File: %s\nError: %s\n\n", err.File, err.Err.Error())
	}

	return nil
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

type ProcessErr struct {
	File string
	Err  error
}
