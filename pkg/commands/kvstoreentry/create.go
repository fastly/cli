package kvstoreentry

import (
	"bufio"
	"bytes"
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

func (c *CreateCommand) ProcessFile(in io.Reader, out io.Writer) error {
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

func (c *CreateCommand) ProcessDir(in io.Reader, out io.Writer) error {
	files, err := os.ReadDir(c.dirPath)
	if err != nil {
		return err
	}

	var fileBatch []string
	var wg sync.WaitGroup

	for _, file := range files {
		if !file.IsDir() {
			fileBatch = append(fileBatch, filepath.Join(c.dirPath, file.Name()))
			if len(fileBatch) == c.dirBatchSize {
				wg.Add(1)
				go processBatch(fileBatch, &wg, c.CallEndpoint)
				fileBatch = nil
			}
		}
	}

	if len(fileBatch) > 0 {
		wg.Add(1)
		go processBatch(fileBatch, &wg, c.CallEndpoint)
	}

	wg.Wait()
	text.Success(out, "Inserted keys into KV Store")
	return nil
}

func processBatch(filePaths []string, wg *sync.WaitGroup, callEndpoint func(in io.Reader) error) {
	defer wg.Done()

	var payload bytes.Buffer
	template := `{"key": "%s", "value": "%s"}`

	for _, filePath := range filePaths {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("error reading file:", err)
			continue
		}

		_, filename := filepath.Split(filePath)
		filenameNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		// We need to strip a trailing newline from the file content (if present).
		if len(fileContent) > 0 && fileContent[len(fileContent)-1] == '\n' {
			fileContent = fileContent[:len(fileContent)-1]
		}

		payload.WriteString(fmt.Sprintf(template, filenameNoExt, fileContent))
		payload.WriteByte('\n')
	}

	fmt.Println(payload.String())
	_ = callEndpoint(&payload)
}

func (c *CreateCommand) CallEndpoint(in io.Reader) error {
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
		host, path, http.MethodPut, token, io.TeeReader(body, os.Stdout), c.Globals.HTTPClient,
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
