package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
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
	c.CmdClause = parent.Command("switch", "Switch user profile (deprecated: use 'fastly auth use' instead)")
	c.CmdClause.Arg("profile", "Profile to switch to").Short('p').Required().StringVar(&c.profile)
	return &c
}

// Exec invokes the application logic for the command.
func (c *SwitchCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth use' instead.\n\n")

	at := c.Globals.Config.GetAuthToken(c.profile)
	if at == nil {
		err := fmt.Errorf("the profile '%s' does not exist", c.profile)
		c.Globals.ErrLog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	if at.Type == config.AuthTokenTypeSSO {
		if err := authcmd.RunSSOWithTokenName(in, out, c.Globals, false, false, c.profile); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		if err := c.Globals.Config.SetDefaultAuthToken(c.profile); err != nil {
			return err
		}
		if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
			return fmt.Errorf("error saving config file: %w", err)
		}
		text.Success(out, "\nProfile switched to '%s'", c.profile)
		return nil
	}

	if err := c.Globals.Config.SetDefaultAuthToken(c.profile); err != nil {
		c.Globals.ErrLog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fsterr.ProfileRemediation(),
		}
	}

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
