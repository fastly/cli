package profile

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// SwitchCommand represents a Kingpin command.
type SwitchCommand struct {
	cmd.Base

	profile string
}

// NewSwitchCommand returns a usable command registered under the parent.
func NewSwitchCommand(parent cmd.Registerer, globals *config.Data) *SwitchCommand {
	var c SwitchCommand
	c.Globals = globals
	c.CmdClause = parent.Command("switch", "Switch user profile")
	c.CmdClause.Arg("profile", "Profile to switch to").Short('p').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SwitchCommand) Exec(in io.Reader, out io.Writer) error {
	var ok bool

	if c.Globals.File.Profiles, ok = profile.Set(c.profile, c.Globals.File.Profiles); !ok {
		err := errors.New(profile.DoesNotExist)
		c.Globals.ErrLog.Add(err)
		return err
	}

	if err := c.Globals.File.Write(c.Globals.Path); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	text.Success(out, "Profile switched to '%s'", c.profile)
	return nil
}
