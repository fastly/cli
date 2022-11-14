package privatekey

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a TLS certificate").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("key", "The contents of the private key. Must be a PEM-formatted key").Required().StringVar(&c.key)
	c.CmdClause.Flag("name", "A customizable name for your private key").Required().StringVar(&c.name)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	key      string
	manifest manifest.Data
	name     string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

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
func (c *CreateCommand) constructInput() *fastly.CreatePrivateKeyInput {
	var input fastly.CreatePrivateKeyInput

	input.Key = c.key
	input.Name = c.name

	return &input
}
