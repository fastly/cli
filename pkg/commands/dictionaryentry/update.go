package dictionaryentry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	argparser.Base

	Input       fastly.UpdateDictionaryItemInput
	InputBatch  fastly.BatchModifyDictionaryItemsInput
	file        argparser.OptionalString
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
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("value", "Dictionary item value").StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	d, err := c.Globals.APIClient.UpdateDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated dictionary item (service %s)\n\n", fastly.ToValue(d.ServiceID))
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

	err = c.Globals.APIClient.BatchModifyDictionaryItems(&c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.InputBatch.Items), c.Input.DictionaryID, c.InputBatch.ServiceID)
	return nil
}
