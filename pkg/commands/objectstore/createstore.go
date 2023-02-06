package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateStoreCommand calls the Fastly API to create an object store.
type CreateStoreCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.CreateObjectStoreInput
}

// NewCreateStoreCommand returns a usable command registered under the parent.
func NewCreateStoreCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateStoreCommand {
	c := CreateStoreCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("create", "Create an object store")
	c.CmdClause.Flag("name", "Name of Object Store").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateStoreCommand) Exec(_ io.Reader, out io.Writer) error {
	d, err := c.Globals.APIClient.CreateObjectStore(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created object store %s", d.Name)
	return nil
}
