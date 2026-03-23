package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand represents a Kingpin command.
type DeleteCommand struct {
	argparser.Base

	profile string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = g
	c.CmdClause = parent.Command("delete", "Delete user profile (deprecated: use 'fastly auth delete' instead)")
	c.CmdClause.Arg("profile", "Profile to delete").Short('x').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.Globals.Flags.Quiet {
		text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth delete' instead.\n\n")
	}

	if !c.Globals.Config.DeleteAuthToken(c.profile) {
		return fmt.Errorf("the specified profile does not exist")
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return err
	}
	if c.Globals.Verbose() {
		text.Break(out)
	}
	text.Success(out, "Profile '%s' deleted", c.profile)

	if c.Globals.Config.Auth.Default == "" && len(c.Globals.Config.Auth.Tokens) > 0 {
		text.Break(out)
		text.Warning(out, "At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>` and ensure the profile is set to be the default.")
	}
	return nil
}
