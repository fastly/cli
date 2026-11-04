package privatekey

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	Key      string `help:"The contents of the private key. Must be a PEM-formatted key, mutually exclusive with --key-path"`
	KeyPath  string `help:"Filepath to a PEM-formatted key, mutually exclusive with --key"`
	NameFlag string `name:"name" required:"" help:"A customizable name for your private key"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.CreatePrivateKey(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Private Key Name": c.NameFlag,
		})
		return err
	}

	text.Success(out, "Created TLS Private Key '%s'", r.Name)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() (*fastly.CreatePrivateKeyInput, error) {
	var input fastly.CreatePrivateKeyInput

	if c.KeyPath == "" && c.Key == "" {
		return nil, fmt.Errorf("neither --key-path or --key provided, one must be provided")
	}

	if c.KeyPath != "" && c.Key != "" {
		return nil, fmt.Errorf("--key-path and --key provided, only one can be specified")
	}

	if c.Key != "" {
		input.Key = c.Key
	}

	if c.KeyPath != "" {
		path, err := filepath.Abs(c.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing key-path '%s': %q", c.KeyPath, err)
		}

		data, err := os.ReadFile(path) // #nosec
		if err != nil {
			return nil, fmt.Errorf("error reading key-path '%s': %q", c.KeyPath, err)
		}

		input.Key = string(data)
	}

	input.Name = c.NameFlag

	return &input, nil
}
