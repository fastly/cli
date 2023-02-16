package configstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("describe", "Retrieve a single config store item").Alias("get")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "key",
		Short:       'k',
		Description: "Item name",
		Dst:         &c.input.Key,
		Required:    true,
	})
	c.RegisterFlag(cmd.StoreIDFlag(&c.input.StoreID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	input    fastly.GetConfigStoreItemInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := cmd.Globals.APIClient.GetConfigStoreItem(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.PrintConfigStoreItem(out, "", o)

	return nil
}
