package dictionaryentry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	argparser.Base

	Input       fastly.UpdateDictionaryItemInput
	InputBatch  fastly.BatchModifyDictionaryItemsInput
	file        argparser.OptionalString
	perItem     bool
	serviceName argparser.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update or insert an item on a Fastly edge dictionary")

	// Required.
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)

	// Optional.
	c.CmdClause.Flag("file", "Batch update json file").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("per-item", "With --file, send one request per item and report each failure. Successful changes still apply").BoolVar(&c.perItem)
	c.CmdClause.Flag("key", "Dictionary item key").StringVar(&c.Input.ItemKey)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("value", "Dictionary item value").StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.perItem && !c.file.WasSet {
		return fmt.Errorf("--per-item requires --file")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
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

	d, err := c.Globals.APIClient.UpdateDictionaryItem(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated dictionary item (service %s)\n\n", fastly.ToValue(d.ServiceID))
	text.PrintDictionaryItem(out, "", d)
	return nil
}

func (c *UpdateCommand) batchModify(out io.Writer) error {
	jsonBytes, err := os.ReadFile(c.file.Value)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &c.InputBatch)
	if err != nil {
		return err
	}

	if len(c.InputBatch.Items) == 0 {
		return fmt.Errorf("item key not found in file %s", c.file.Value)
	}

	if c.perItem {
		if err := c.modifyPerItem(); err != nil {
			return err
		}
	} else if err := c.Globals.APIClient.BatchModifyDictionaryItems(context.TODO(), &c.InputBatch); err != nil {
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.InputBatch.Items), c.Input.DictionaryID, c.InputBatch.ServiceID)
	return nil
}

func (c *UpdateCommand) modifyPerItem() error {
	for index, item := range c.InputBatch.Items {
		if item == nil || fastly.ToValue(item.ItemKey) == "" {
			return fmt.Errorf("item %d in %s is missing an item key", index+1, c.file.Value)
		}
	}

	input := c.InputBatch
	var failures []error
	for index, item := range c.InputBatch.Items {
		input.Items = c.InputBatch.Items[index : index+1]
		if err := c.Globals.APIClient.BatchModifyDictionaryItems(context.TODO(), &input); err != nil {
			failures = append(failures, fmt.Errorf("item %d/%d (key %q): %w", index+1, len(c.InputBatch.Items), fastly.ToValue(item.ItemKey), err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d of %d dictionary item requests failed:\n%w", len(failures), len(c.InputBatch.Items), errors.Join(failures...))
	}
	return nil
}
