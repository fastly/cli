package edgedictionaryitem

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.UpdateDictionaryItemInput
}

// NewUpdateCommand returns a usable command registered under the parent.
//
// TODO(integralist) update to not use common.OptionalString once we have a
// new Go-Fastly release that modifies UpdateDictionaryItemInput so that the
// ItemValue is no longer optional.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update or insert an item on a Fastly edge dictionary")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)
	c.CmdClause.Flag("value", "Dictionary item value").Required().StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	d, err := c.Globals.Client.UpdateDictionaryItem(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated dictionary item (service %s)", d.ServiceID)
	text.Break(out)
	text.PrintDictionaryItem(out, "", d)
	return nil
}
