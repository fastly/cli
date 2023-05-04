package kvstoreentry

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"

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
	filePath string
	manifest manifest.Data
	stdin    bool

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
	c.CmdClause.Flag("file", "Path to a file containing individual JSON objects separated by new-line delimiter").StringVar(&c.filePath)
	c.CmdClause.Flag("key-name", "Key name").Short('k').StringVar(&c.Input.Key)
	c.CmdClause.Flag("stdin", "Read new-line separated JSON stream via STDIN").BoolVar(&c.stdin)
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("value", "Value").StringVar(&c.Input.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.stdin && c.filePath != "" {
		return fsterr.ErrInvalidStdinFileCombo
	}

	if c.stdin {
		return c.ProcessStdin(in, out)
	}

	if c.filePath != "" {
		return c.ProcessFile(in, out)
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
