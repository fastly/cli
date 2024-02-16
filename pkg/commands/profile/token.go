package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/profile"
	"github.com/fastly/cli/v10/pkg/text"
)

// TokenCommand represents a Kingpin command.
type TokenCommand struct {
	argparser.Base
	profile string
}

// NewTokenCommand returns a new command registered in the parent.
func NewTokenCommand(parent argparser.Registerer, g *global.Data) *TokenCommand {
	var c TokenCommand
	c.Globals = g
	c.CmdClause = parent.Command("token", "Print API token (defaults to the 'active' profile)")
	c.CmdClause.Arg("profile", "Print API token for the named profile").Short('p').StringVar(&c.profile)
	return &c
}

// Exec implements the command interface.
func (c *TokenCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	var p string
	if c.profile != "" {
		p = c.profile
	}
	if c.Globals.Flags.Profile != "" {
		p = c.Globals.Flags.Profile
		// NOTE: If global --profile is set, it take precedence over 'profile' arg.
		// It's unlikely someone will provide both, but we'll code defensively.
	}

	if p != "" {
		if p := profile.Get(p, c.Globals.Config.Profiles); p != nil {
			text.Output(out, p.Token)
			return nil
		}
		msg := fmt.Sprintf(profile.DoesNotExist, p)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	// If no 'profile' arg or global --profile, then we'll use 'active' profile.
	if _, p := profile.Default(c.Globals.Config.Profiles); p != nil {
		text.Output(out, p.Token)
		return nil
	}
	return fsterr.RemediationError{
		Inner:       fmt.Errorf("no profiles available"),
		Remediation: fsterr.ProfileRemediation,
	}
}
