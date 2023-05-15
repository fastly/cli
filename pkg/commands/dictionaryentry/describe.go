package dictionaryentry

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// DescribeCommand calls the Fastly API to describe a dictionary item.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest    manifest.Data
	Input       fastly.GetDictionaryItemInput
	serviceName cmd.OptionalServiceNameID
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly edge dictionary item").Alias("get")

	// Required.
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
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
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	o, err := c.Globals.APIClient.GetDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", c.Input.ServiceID)
	}
	text.PrintDictionaryItem(out, "", o)
	return nil
}
