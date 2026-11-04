package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// SwitchCommand represents a Kingpin command.
type SwitchCommand struct {
	argparser.Base

	Profile string `arg:"" required:"" help:"Profile to switch to"`
}

// Exec invokes the application logic for the command.
func (c *SwitchCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.Globals.Flags.Quiet {
		text.Deprecated("This command will be removed in a future release. Use 'fastly auth use' instead.\n\n")
	}

	at := c.Globals.Config.GetAuthToken(c.Profile)
	if at == nil {
		err := fmt.Errorf("the profile '%s' does not exist", c.Profile)
		c.Globals.ErrLog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	if at.Type == config.AuthTokenTypeSSO {
		if err := authcmd.RunSSOWithTokenName(in, out, c.Globals, false, false, c.Profile); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		if err := c.Globals.Config.SetDefaultAuthToken(c.Profile); err != nil {
			return err
		}
		if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
			return fmt.Errorf("error saving config file: %w", err)
		}
		text.Success(out, "\nProfile switched to '%s'", c.Profile)
		return nil
	}

	if err := c.Globals.Config.SetDefaultAuthToken(c.Profile); err != nil {
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
	text.Success(out, "Profile switched to '%s'", c.Profile)
	return nil
}
