package dictionaryentry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a dictionary item.
type UpdateCommand struct {
	argparser.Base

	Input      fastly.UpdateDictionaryItemInput       `kong:"-"`
	InputBatch fastly.BatchModifyDictionaryItemsInput `kong:"-"`

	// Required.
	DictionaryID string `required:"" help:"Dictionary ID"`

	// Optional.
	File        argparser.OptionalString        `help:"Batch update json file"`
	Key         string                          `help:"Dictionary item key"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Value       string                          `help:"Dictionary item value"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.Input.DictionaryID = c.DictionaryID
	c.Input.ItemKey = c.Key
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.Input.ItemValue = c.Value
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID
	c.InputBatch.ServiceID = serviceID
	c.InputBatch.DictionaryID = c.Input.DictionaryID

	if c.File.WasSet {
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
	jsonFile, err := os.Open(c.File.Value)
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
		return fmt.Errorf("item key not found in file %s", c.File.Value)
	}

	err = c.Globals.APIClient.BatchModifyDictionaryItems(context.TODO(), &c.InputBatch)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.InputBatch.Items), c.Input.DictionaryID, c.InputBatch.ServiceID)
	return nil
}
