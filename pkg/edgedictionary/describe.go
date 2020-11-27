package edgedictionary

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// DescribeCommand calls the Fastly API to describe a service.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetDictionaryInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly edge dictionary").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	dictionary, err := c.Globals.Client.GetDictionary(&c.Input)
	if err != nil {
		return err
	}

	text.Output(out, "Service ID: %s", dictionary.ServiceID)
	text.Output(out, "Version: %d", dictionary.ServiceVersion)
	text.PrintDictionary(out, "", dictionary)

	if c.Globals.Verbose() {
		infoInput := fastly.GetDictionaryInfoInput{
			ServiceID:      c.Input.ServiceID,
			ServiceVersion: c.Input.ServiceVersion,
			ID:             c.manifest.Flag.ServiceID,
		}
		info, err := c.Globals.Client.GetDictionaryInfo(&infoInput)
		if err != nil {
			return err
		}
		text.Output(out, "Digest: %s", info.Digest)
		text.Output(out, "Item Count: %d", info.ItemCount)

		itemInput := fastly.ListDictionaryItemsInput{
			ServiceID:    c.Input.ServiceID,
			DictionaryID: dictionary.ID,
		}
		items, err := c.Globals.Client.ListDictionaryItems(&itemInput)
		if err != nil {
			return err
		}
		for i, item := range items {
			text.Output(out, "Item %d/%d:", i+1, len(items))
			text.PrintDictionaryItemKV(out, "	", item)
		}
	}

	return nil
}
