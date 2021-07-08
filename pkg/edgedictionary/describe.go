package edgedictionary

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a dictionary.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetDictionaryInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly edge dictionary").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	dictionary, err := c.Globals.Client.GetDictionary(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Output(out, "Service ID: %s", dictionary.ServiceID)
	text.Output(out, "Version: %d", dictionary.ServiceVersion)
	text.PrintDictionary(out, "", dictionary)

	if c.Globals.Verbose() {
		infoInput := fastly.GetDictionaryInfoInput{
			ServiceID:      c.Input.ServiceID,
			ServiceVersion: c.Input.ServiceVersion,
			ID:             dictionary.ID,
		}
		info, err := c.Globals.Client.GetDictionaryInfo(&infoInput)
		if err != nil {
			c.Globals.ErrLog.Add(err)
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
			c.Globals.ErrLog.Add(err)
			return err
		}
		for i, item := range items {
			text.Output(out, "Item %d/%d:", i+1, len(items))
			text.PrintDictionaryItemKV(out, "	", item)
		}
	}

	return nil
}
