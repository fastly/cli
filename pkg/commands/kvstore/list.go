package kvstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list the available kv stores.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest manifest.Data
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("list", "List kv stores")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var cursor string

	for {
		o, err := c.Globals.APIClient.ListKVStores(&fastly.ListKVStoresInput{
			Cursor: cursor,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			// No pagination prompt w/ JSON output.
			// FIXME: This should be fixed here and for Secrets Store.
			return err
		}

		if o != nil {
			for _, o := range o.Data {
				// avoid gosec loop aliasing check :/
				o := o
				text.PrintKVStore(out, "", &o)
			}
			if cur, ok := o.Meta["next_cursor"]; ok && cur != "" && cur != cursor {
				// Check if 'out' is interactive before prompting.
				if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
					text.Break(out)
					printNext, err := text.AskYesNo(out, "Print next page [yes/no]: ", in)
					if err != nil {
						return err
					}
					if printNext {
						cursor = cur
						continue
					}
				} else {
					// Otherwise if non-interactive or auto-yes, then load all data.
					cursor = cur
					continue
				}
			}
		}

		return nil
	}
}
