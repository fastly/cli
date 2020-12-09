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

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.UpdateDictionaryItemInput

	itemvalue common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update or insert an item on a Fastly edge dictionary")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)
	c.CmdClause.Flag("value", "Dictionary item value").Required().Action(c.itemvalue.Set).StringVar(&c.itemvalue.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	c.Input.ItemValue = &c.itemvalue.Value

	dictionary, err := c.Globals.Client.UpdateDictionaryItem(&c.Input)
	if err != nil {
		return err
	}

	text.Output(out, "Service ID: %s", c.Input.ServiceID)
	text.PrintDictionaryItem(out, "", dictionary)
	return nil
}
