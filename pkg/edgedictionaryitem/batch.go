package edgedictionaryitem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// BatchCommand calls the Fastly API to batch update a dictionary.
type BatchCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.BatchModifyDictionaryItemsInput

	file cmd.OptionalString
}

// NewBatchCommand returns a usable command registered under the parent.
func NewBatchCommand(parent cmd.Registerer, globals *config.Data) *BatchCommand {
	var c BatchCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("batchmodify", "Update multiple items in a Fastly edge dictionary")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("file", "Batch update json file").Required().Action(c.file.Set).StringVar(&c.file.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *BatchCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

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

	err = json.Unmarshal(jsonBytes, &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if len(c.Input.Items) == 0 {
		return fmt.Errorf("item key not found in file %s", c.file.Value)
	}

	err = c.Globals.Client.BatchModifyDictionaryItems(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Made %d modifications of Dictionary %s on service %s", len(c.Input.Items), c.Input.DictionaryID, c.Input.ServiceID)
	return nil
}
