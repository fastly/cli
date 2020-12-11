package edgedictionaryitem

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// CreateCommand calls the Fastly API to create a dictionary item.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateDictionaryItemInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a new item on a Fastly edge dictionary")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)
	c.CmdClause.Flag("value", "Dictionary item value").Required().StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	_, err := c.Globals.Client.CreateDictionaryItem(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created dictionary item %s (service %s, dictionary %s)", c.Input.ItemKey, c.Input.ServiceID, c.Input.DictionaryID)

	return nil
}
