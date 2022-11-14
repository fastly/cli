package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// ListCommand calls the Fastly API to list the available object stores.
type ListCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListObjectStoresInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Fastly edge config stores")
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	o, err := c.Globals.APIClient.ListObjectStores(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	for _, o := range o.Data {
		// avoid gosec loop aliasing check :/
		o := o
		text.PrintObjectStore(out, "", &o)
	}

	return nil
}
