package edgedictionary

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
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
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	dictionary, err := c.Globals.Client.GetDictionary(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", dictionary.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", dictionary.Version)
	text.PrintDictionary(out, "", dictionary)

	if c.Globals.Verbose() {
		infoInput := fastly.GetDictionaryInfoInput{
			ServiceID: c.Input.Service,
			Version:   c.Input.Version,
			ID:        c.manifest.Flag.ServiceID,
		}
		info, err := c.Globals.Client.GetDictionaryInfo(&infoInput)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Digest: %s\n", info.Digest)
		fmt.Fprintf(out, "Item Count: %d\n", info.ItemCount)

		itemInput := fastly.ListDictionaryItemsInput{
			Service:    c.Input.Service,
			Dictionary: dictionary.ID,
		}
		items, err := c.Globals.Client.ListDictionaryItems(&itemInput)
		if err != nil {
			return err
		}
		for i, item := range items {
			fmt.Fprintf(out, "	Item %d/%d:\n", i+1, len(items))
			text.PrintDictionaryItemKV(out, "		", item)
		}
	}

	return nil
}
