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
	c := ListSecretsCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}

	c.CmdClause = parent.Command("list", "List secrets within a specified store")

	// Required.
	c.RegisterFlag(storeIDFlag(&c.Input.ID)) // --store-id

	// Optional.
	c.RegisterFlag(cursorFlag(&c.Input.Cursor)) // --cursor
	c.RegisterFlagBool(c.jsonFlag())            // --json
	limitFlag(c.CmdClause, &c.Input.Limit)      // --limit

	return &c
}

// ListSecretsCommand calls the Fastly API to list appropriate resources.
type ListSecretsCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.ListSecretsInput
	manifest manifest.Data
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
