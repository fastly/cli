package privatekey

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a TLS private key").Alias("add")
	c.Globals = g

	// Required.
	c.CmdClause.Flag("key", "The contents of the private key. Must be a PEM-formatted key, mutually exclusive with --key-path").StringVar(&c.key)
	c.CmdClause.Flag("key-path", "Filepath to a PEM-formatted key, mutually exclusive with --key").StringVar(&c.keyPath)
	c.CmdClause.Flag("name", "A customizable name for your private key").Required().StringVar(&c.name)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	key     string
	keyPath string
	name    string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.CreatePrivateKey(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Private Key Name": c.name,
		})
		return err
	}

	text.Success(out, "Created TLS Private Key '%s'", r.Name)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() (*fastly.CreatePrivateKeyInput, error) {
	var input fastly.CreatePrivateKeyInput

	if c.keyPath == "" && c.key == "" {
		return nil, fmt.Errorf("neither --key-path or --key provided, one must be provided")
	}

	if c.keyPath != "" && c.key != "" {
		return nil, fmt.Errorf("--key-path and --key provided, only one can be specified")
	}

	if c.key != "" {
		input.Key = c.key
	}

	if c.keyPath != "" {
		path, err := filepath.Abs(c.keyPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing key-path '%s': %q", c.keyPath, err)
		}

		data, err := os.ReadFile(path) // #nosec
		if err != nil {
			return nil, fmt.Errorf("error reading key-path '%s': %q", c.keyPath, err)
		}

		input.Key = string(data)
	}

	input.Name = c.name

	return &input, nil
}
