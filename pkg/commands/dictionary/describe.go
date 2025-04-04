package dictionary

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a dictionary.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.GetDictionaryInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly edge dictionary").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
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
	return &c
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
		ServiceNameFlag:    c.serviceName,
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

	dictionary, err := c.Globals.APIClient.GetDictionary(&c.Input)
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
		info, err = c.Globals.APIClient.GetDictionaryInfo(&infoInput)
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
		items, err = c.Globals.APIClient.ListDictionaryItems(&itemInput)
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
