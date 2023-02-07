package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand represents a Kingpin command.
type DeleteCommand struct {
	cmd.Base

	profile string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = g
	c.CmdClause = parent.Command("delete", "Delete user profile")
	c.CmdClause.Arg("profile", "Profile to delete").Short('x').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if ok := profile.Delete(c.profile, c.Globals.Config.Profiles); ok {
		if err := c.Globals.Config.Write(c.Globals.Path); err != nil {
			return err
		}
		text.Success(out, "Profile '%s' deleted", c.profile)

		if p, _ := profile.Default(c.Globals.Config.Profiles); p == "" && len(c.Globals.Config.Profiles) > 0 {
			text.Warning(out, profile.NoDefaults)
		}
		return nil
	}
	return fmt.Errorf("the specified profile does not exist")
}
