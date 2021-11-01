package dictionaryitem

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// DeleteCommand calls the Fastly API to delete a service.
type DeleteCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteDictionaryItemInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("delete", "Delete an item from a Fastly edge dictionary")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	err := c.Globals.Client.DeleteDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Deleted dictionary item %s (service %s, dicitonary %s)", c.Input.ItemKey, c.Input.ServiceID, c.Input.DictionaryID)
	return nil
}
