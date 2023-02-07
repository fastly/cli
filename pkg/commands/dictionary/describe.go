package dictionary

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// DescribeCommand calls the Fastly API to describe a dictionary.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetDictionaryInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly edge dictionary").Alias("get")

	// required
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	dictionary, err := c.Globals.APIClient.GetDictionary(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	var (
		info  *fastly.DictionaryInfo
		items []*fastly.DictionaryItem
	)
	if c.Globals.Verbose() || c.json {
		infoInput := fastly.GetDictionaryInfoInput{
			ServiceID:      c.Input.ServiceID,
			ServiceVersion: c.Input.ServiceVersion,
			ID:             dictionary.ID,
		}
		info, err = c.Globals.APIClient.GetDictionaryInfo(&infoInput)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
		itemInput := fastly.ListDictionaryItemsInput{
			ServiceID:    c.Input.ServiceID,
			DictionaryID: dictionary.ID,
		}
		items, err = c.Globals.APIClient.ListDictionaryItems(&itemInput)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
	}

	if c.json {
		// NOTE: When not using JSON you have to provide the --verbose flag to get
		// some extra information about the dictionary. When using --json we go
		// ahead and acquire that info and combine it into the JSON output.
		type container struct {
			*fastly.Dictionary
			*fastly.DictionaryInfo
			Items []*fastly.DictionaryItem
		}
		data, err := json.Marshal(&container{Dictionary: dictionary, DictionaryInfo: info, Items: items})
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

	if !c.Globals.Verbose() {
		text.Output(out, "Service ID: %s", dictionary.ServiceID)
	}
	text.Output(out, "Version: %d", dictionary.ServiceVersion)
	text.PrintDictionary(out, "", dictionary)

	if c.Globals.Verbose() {
		text.Output(out, "Digest: %s", info.Digest)
		text.Output(out, "Item Count: %d", info.ItemCount)

		for i, item := range items {
			text.Output(out, "Item %d/%d:", i+1, len(items))
			text.PrintDictionaryItemKV(out, "	", item)
		}
	}

	return nil
}
