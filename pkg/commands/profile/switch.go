package profile

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// SwitchCommand represents a Kingpin command.
type SwitchCommand struct {
	argparser.Base

	profile string
}

// NewSwitchCommand returns a usable command registered under the parent.
func NewSwitchCommand(parent argparser.Registerer, g *global.Data) *SwitchCommand {
	var c SwitchCommand
	c.Globals = g
	c.CmdClause = parent.Command("switch", "Switch user profile")
	c.CmdClause.Arg("profile", "Profile to switch to").Short('p').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SwitchCommand) Exec(_ io.Reader, out io.Writer) error {
	var ok bool

	// We call SetDefault for its side effect of resetting all other profiles to have
	// their Default field set to false.
	p, ok := profile.SetDefault(c.profile, c.Globals.Config.Profiles)
	if !ok {
		msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
		err := errors.New(msg)
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Globals.Config.Profiles = p

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

  if c.Globals.Verbose() {
    text.Break(out)
  }
	text.Success(out, "Profile switched to '%s'", c.profile)
	return nil
}
