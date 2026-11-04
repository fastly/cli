package dictionary

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a dictionary.
type DescribeCommand struct {
	argparser.Base

	Input          fastly.GetDictionaryInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Name of Dictionary"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersionNumber

	dictionary, err := c.Globals.APIClient.GetDictionary(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}
	dictionaryID := fastly.ToValue(dictionary.DictionaryID)

	var (
		info  *fastly.DictionaryInfo
		items []*fastly.DictionaryItem
	)

	if c.Globals.Verbose() || c.JSONOutput.Enabled {
		infoInput := fastly.GetDictionaryInfoInput{
			ServiceID:      c.Input.ServiceID,
			ServiceVersion: c.Input.ServiceVersion,
			DictionaryID:   dictionaryID,
		}
		info, err = c.Globals.APIClient.GetDictionaryInfo(context.TODO(), &infoInput)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
		itemInput := fastly.ListDictionaryItemsInput{
			ServiceID:    c.Input.ServiceID,
			DictionaryID: dictionaryID,
		}
		items, err = c.Globals.APIClient.ListDictionaryItems(context.TODO(), &itemInput)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
	}

	if c.JSONOutput.Enabled {
		// NOTE: When not using JSON you have to provide the --verbose flag to get
		// some extra information about the dictionary. When using --json we go
		// ahead and acquire that info and combine it into the JSON output.
		type container struct {
			*fastly.Dictionary
			*fastly.DictionaryInfo
			Items []*fastly.DictionaryItem
		}

		o := &container{Dictionary: dictionary, DictionaryInfo: info, Items: items}

		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}
	}

	if !c.Globals.Verbose() {
		text.Output(out, "Service ID: %s", fastly.ToValue(dictionary.ServiceID))
	}
	text.Output(out, "Version: %d", fastly.ToValue(dictionary.ServiceVersion))
	text.PrintDictionary(out, "", dictionary)

	if c.Globals.Verbose() {
		text.Output(out, "Digest: %s", fastly.ToValue(info.Digest))
		text.Output(out, "Item Count: %d", fastly.ToValue(info.ItemCount))

		for i, item := range items {
			text.Output(out, "Item %d/%d:", i+1, len(items))
			text.PrintDictionaryItemKV(out, "	", item)
		}
	}

	return nil
}
