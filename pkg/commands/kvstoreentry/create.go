package kvstoreentry

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	c.CmdClause.Flag("dir-batch-size", "Files will be processed in batches").Default("10").IntVar(&c.dirBatchSize)
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
	dirBatchSize int
	dirPath      string
	filePath     string
	manifest     manifest.Data
	stdin        bool

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
		return c.ProcessFile(in, out)
	}

	if c.dirPath != "" {
		return c.ProcessDir(in, out)
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

	if err := c.CallEndpoint(in); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func (c *CreateCommand) ProcessFile(_ io.Reader, out io.Writer) error {
	f, err := os.Open(c.filePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if err := c.CallEndpoint(f); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func (c *CreateCommand) ProcessDir(_ io.Reader, out io.Writer) error {
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
	processing := fmt.Sprintf("Processing %d files", len(files))
	msg := "%s (files read: %d, api calls made: %d)"
	spinner.Message(fmt.Sprintf(msg, processing, 0, 0) + "...")

	base := filepath.Base(path)
	errors := []ProcessErr{}
	errMessages := make(chan ProcessErr)
	fileRead := make(chan bool)
	apiCall := make(chan bool)

	var (
		apiCalls  int
		filesRead int
		fileBatch []string
		wg        sync.WaitGroup
	)

	go func() {
		for msg := range errMessages {
			fmt.Println("received message to errMessages channel")
			errors = append(errors, msg)
		}
	}()

	go func() {
		for ok := range fileRead {
			if ok {
				fmt.Println("received message to fileRead channel")
				filesRead++
				spinner.Message(fmt.Sprintf(msg, processing, filesRead, apiCalls) + "...")
			}
		}
	}()

	go func() {
		for ok := range apiCall {
			if ok {
				fmt.Println("received message to apiCall channel")
				apiCalls++
				spinner.Message(fmt.Sprintf(msg, processing, filesRead, apiCalls) + "...")
			}
		}
	}()

	for _, file := range files {
		if !file.IsDir() {
			fileBatch = append(fileBatch, filepath.Join(c.dirPath, file.Name()))
			if len(fileBatch) == c.dirBatchSize {
				wg.Add(1)
				go processBatch(base, fileBatch, &wg, c.CallEndpoint, &errMessages, &fileRead, &apiCall)
				fileBatch = nil
			}
		}
	}

	if len(fileBatch) > 0 {
		wg.Add(1)
		go processBatch(base, fileBatch, &wg, c.CallEndpoint, &errMessages, &fileRead, &apiCall)
	}

	wg.Wait()

	spinner.StopMessage(fmt.Sprintf(msg, processing, filesRead, apiCalls))
	err = spinner.Stop()
	if err != nil {
		return err
	}

	if len(errors) == 0 {
		text.Success(out, "Inserted %d keys into KV Store", len(files))
		return nil
	}

	// NOTE: We can't be accurate because a batch of files might fail.
	// The API response doesn't indicate what files from the batch failed.
	// Just that the batch failed to upload.
	text.Error(out, "Inserted (approx) %d keys into KV Store", len(files)-len(errors))

	for _, err := range errors {
		fmt.Printf("File: %s\nError: %s\n\n", err.File, err.Err.Error())
	}

	return nil
}

func (c *CreateCommand) CallEndpoint(in io.Reader) error {
	fmt.Printf("%+v\n", "api call started")
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

	fmt.Printf("%+v\n", "api call finished")
	return nil
}

func processBatch(
	base string,
	filePaths []string,
	wg *sync.WaitGroup,
	callEndpoint func(in io.Reader) error,
	errMessages *chan ProcessErr,
	fileRead *chan bool,
	apiCall *chan bool,
) {
	defer wg.Done()

	var payload bytes.Buffer
	template := `{"key": "%s", "value": "%s"}`

	for _, filePath := range filePaths {
		// gosec flagged this:
		// G304 (CWE-22): Potential file inclusion via variable
		// #nosec
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("send to errMessages channel (file read error)")
			*errMessages <- ProcessErr{
				File: filePath,
				Err:  err,
			}
			continue
		}

		dir, filename := filepath.Split(filePath)
		index := strings.Index(dir, base)
		filename = filepath.Join(dir[index:], filename)

		_, _ = payload.WriteString(fmt.Sprintf(template, filename, base64.StdEncoding.EncodeToString(fileContent)))
		_ = payload.WriteByte('\n')
		fmt.Println("send to fileRead channel")
		*fileRead <- true
	}

	err := callEndpoint(&payload)
	fmt.Println("send to apiCall channel")
	*apiCall <- true
	if err != nil {
		fmt.Println("send to errMessages channel (api call error)")
		*errMessages <- ProcessErr{
			File: "Batch",
			Err:  err,
		}
	}
}

type ProcessErr struct {
	File string
	Err  error
}
