package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/sso"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// SwitchCommand represents a Kingpin command.
type SwitchCommand struct {
	argparser.Base

	profile string
	ssoCmd  *sso.RootCommand
}

// NewSwitchCommand returns a usable command registered under the parent.
func NewSwitchCommand(parent argparser.Registerer, g *global.Data, ssoCmd *sso.RootCommand) *SwitchCommand {
	var c SwitchCommand
	c.Globals = g
	c.ssoCmd = ssoCmd
	c.CmdClause = parent.Command("switch", "Switch user profile")
	c.CmdClause.Arg("profile", "Profile to switch to").Short('p').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SwitchCommand) Exec(in io.Reader, out io.Writer) error {
	var ok bool
	errNoProfile := fmt.Errorf(profile.DoesNotExist, c.profile)

	// We get the named profile to check if it's an SSO-based profile.
	// If we're switching to an SSO-based profile, then we need to re-auth.
	p := profile.Get(c.profile, c.Globals.Config.Profiles)
	if p == nil {
		c.Globals.ErrLog.Add(errNoProfile)
		return errNoProfile
	}
	if isSSOToken(p) {
		// IMPORTANT: We need to set profile fields for `sso` command.
		//
		// This is so the `sso` command will use this information to trigger the
		// correct authentication flow.
		c.ssoCmd.InvokedFromProfileSwitch = true
		c.ssoCmd.ProfileSwitchName = c.profile
		c.ssoCmd.ProfileDefault = true

		err := c.ssoCmd.Exec(in, out)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		text.Success(out, "\nProfile switched to '%s'", c.profile)
		return nil
	}

	// We call SetDefault for its side effect of resetting all other profiles to have
	// their Default field set to false.
	ps, ok := profile.SetDefault(c.profile, c.Globals.Config.Profiles)
	if !ok {
		c.Globals.ErrLog.Add(errNoProfile)
		return errNoProfile
	}

	c.Globals.Config.Profiles = ps

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	text.Success(out, "Profile switched to '%s'", c.profile)
	return nil
}
