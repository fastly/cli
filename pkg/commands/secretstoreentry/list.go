package secretstoreentry

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

	c.CmdClause = parent.Command("list", "List secrets within a specified store")

	// Required.
	c.RegisterFlag(cmd.StoreIDFlag(&c.Input.ID)) // --store-id

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

	Input    fastly.ListSecretsInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	for {
		o, err := c.Globals.APIClient.ListSecrets(&c.Input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		text.PrintSecretsTbl(out, o)

		if o != nil && o.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}
