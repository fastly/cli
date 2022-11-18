package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewListStoresCommand returns a usable command registered under the parent.
func NewListStoresCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListStoresCommand {
	var c ListStoresCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List secret stores")
	c.RegisterFlag(cursorFlag(&c.Input.Cursor))
	limitFlag(c.CmdClause, &c.Input.Limit)
	c.RegisterFlagBool(c.jsonFlag())
	return &c
}

// ListStoresCommand calls the Fastly API to list the available secret stores.
type ListStoresCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListSecretStoresInput
	jsonOutput
}

// Exec invokes the application logic for the command.
func (cmd *ListStoresCommand) Exec(in io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
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
			if !cmd.Globals.Flag.NonInteractive && text.IsTTY(out) {
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
