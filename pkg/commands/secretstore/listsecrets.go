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

// NewListSecretsCommand returns a usable command registered under the parent.
func NewListSecretsCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListSecretsCommand {
	var c ListSecretsCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List secrets")
	c.RegisterFlag(storeIDFlag(&c.Input.ID))
	c.RegisterFlag(cursorFlag(&c.Input.Cursor))
	limitFlag(c.CmdClause, &c.Input.Limit)
	c.RegisterFlagBool(c.jsonFlag())
	return &c
}

// ListSecretsCommand calls the Fastly API to list the available secret stores.
type ListSecretsCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListSecretsInput
	jsonOutput
}

// Exec invokes the application logic for the command.
func (cmd *ListSecretsCommand) Exec(in io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	for {
		o, err := cmd.Globals.APIClient.ListSecrets(&cmd.Input)
		if err != nil {
			cmd.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := cmd.WriteJSON(out, o); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		text.PrintSecretsTbl(out, o)

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
