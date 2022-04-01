package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand represents a Kingpin command.
type DeleteCommand struct {
	cmd.Base

	profile string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.CmdClause = parent.Command("delete", "Delete user profile")
	c.CmdClause.Arg("profile", "Profile to delete").Short('x').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if ok := profile.Delete(c.profile, c.Globals.File.Profiles); ok {
		if err := c.Globals.File.Write(c.Globals.Path); err != nil {
			return err
		}
		text.Success(out, "The profile '%s' was deleted.", c.profile)

		if profile, _ := profile.Default(c.Globals.File.Profiles); profile == "" && len(c.Globals.File.Profiles) > 0 {
			text.Warning(out, "At least one account profile should be set as the 'default'. Run `fastly configure --profile <NAME>`.")
		}
		return nil
	}
	return fmt.Errorf("the specified profile does not exist")
}
