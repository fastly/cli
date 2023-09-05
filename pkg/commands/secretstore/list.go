package secretstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("list", "List secret stores")

	// Optional.
	c.RegisterFlag(cmd.CursorFlag(&c.Input.Cursor))  // --cursor
	c.RegisterFlagBool(c.JSONFlag())                 // --json
	c.RegisterFlagInt(cmd.LimitFlag(&c.Input.Limit)) // --limit

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	// NOTE: API returns 10 items even when --limit is set to smaller.
	Input    fastly.ListSecretStoresInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var data []fastly.SecretStore

	for {
		o, err := c.Globals.APIClient.ListSecretStores(&c.Input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if o != nil {
			data = append(data, o.Data...)

			if c.JSONOutput.Enabled || c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes {
				if o.Meta.NextCursor != "" {
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
				break
			}

			text.PrintSecretStoresTbl(out, o.Data)

			if o.Meta.NextCursor != "" {
				text.Break(out)
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					text.Break(out)
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
			}
		}

		break
	}

	ok, err := c.WriteJSON(out, data)
	if err != nil {
		return err
	}

	// Only print output here if we've not already printed JSON.
	// And only if we're non interactive.
	// Otherwise interactive mode would have displayed each page of data.
	if !ok && (c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes) {
		text.PrintSecretStoresTbl(out, data)
	}
	return nil
}
