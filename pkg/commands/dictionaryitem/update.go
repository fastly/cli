package dictionaryitem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	cmd.Base

	Input      fastly.UpdateDictionaryItemInput
	InputBatch fastly.BatchModifyDictionaryItemsInput
	file       cmd.OptionalString
	manifest   manifest.Data
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update or insert an item on a Fastly edge dictionary")
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("file", "Batch update json file").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("key", "Dictionary item key").StringVar(&c.Input.ItemKey)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("value", "Dictionary item value").StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID
	c.InputBatch.ServiceID = serviceID
	c.InputBatch.DictionaryID = c.Input.DictionaryID

	if c.file.WasSet {
		err := c.batchModify(out)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		return nil
	}

	if c.Input.ItemKey == "" || c.Input.ItemValue == "" {
		return fmt.Errorf("an empty value is not allowed for either the '--key' or '--value' flags")
	}

	d, err := c.Globals.Client.UpdateDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated dictionary item (service %s)", d.ServiceID)
	text.Break(out)
	text.PrintDictionaryItem(out, "", d)
	return nil
}

func (c *UpdateCommand) batchModify(out io.Writer) error {
	jsonFile, err := os.Open(c.file.Value)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	err = json.Unmarshal(jsonBytes, &c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if len(c.InputBatch.Items) == 0 {
		return fmt.Errorf("item key not found in file %s", c.file.Value)
	}

	err = c.Globals.Client.BatchModifyDictionaryItems(&c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.InputBatch.Items), c.Input.DictionaryID, c.InputBatch.ServiceID)
	return nil
}
