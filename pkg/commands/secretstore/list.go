package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
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

	Input    fastly.ListSecretStoresInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	for {
		o, err := cmd.Globals.APIClient.ListSecretStores(&cmd.Input)
		if err != nil {
			cmd.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := cmd.WriteJSON(out, o); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		text.PrintSecretStoresTbl(out, o)

		if o != nil && o.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !cmd.Globals.Flags.NonInteractive && !cmd.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [yes/no]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					cmd.Input.Cursor = o.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}
