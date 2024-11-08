package profile

import (
	"fmt"
	"io"
	"time"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	fsttime "github.com/fastly/cli/pkg/time"
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

// By default tokens must be valid for at least 5 minutes to be
// considered valid.
const defaultTokenTTL time.Duration = 5 * time.Minute

// Exec implements the command interface.
func (c *TokenCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	var name string
	if c.profile != "" {
		name = c.profile
	}
	if c.Globals.Flags.Profile != "" {
		name = c.Globals.Flags.Profile
		// NOTE: If global --profile is set, it take precedence over 'profile' arg.
		// It's unlikely someone will provide both, but we'll code defensively.
	}

	if name != "" {
		if p := profile.Get(name, c.Globals.Config.Profiles); p != nil {
			if err = checkTokenValidity(name, p, defaultTokenTTL); err != nil {
				return err
			}
			text.Output(out, p.Token)
			return nil
		}
		msg := fmt.Sprintf(profile.DoesNotExist, name)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	// If no 'profile' arg or global --profile, then we'll use 'active' profile.
	if name, p := profile.Default(c.Globals.Config.Profiles); p != nil {
		if err = checkTokenValidity(name, p, defaultTokenTTL); err != nil {
			return err
		}
		text.Output(out, p.Token)
		return nil
	}
	return fsterr.RemediationError{
		Inner:       fmt.Errorf("no profiles available"),
		Remediation: fsterr.ProfileRemediation,
	}
}

func checkTokenValidity(profileName string, p *config.Profile, ttl time.Duration) (err error) {
	var msg string
	expiry := time.Unix(p.RefreshTokenCreated, 0).Add(time.Duration(p.RefreshTokenTTL) * time.Second)

	if expiry.After(time.Now().Add(ttl)) {
		return nil
	} else if expiry.Before(time.Now()) {
		msg = fmt.Sprintf(profile.TokenExpired, profileName, expiry.UTC().Format(fsttime.Format))
	} else {
		msg = fmt.Sprintf(profile.TokenWillExpire, profileName, expiry.UTC().Format(fsttime.Format))
	}

	return fsterr.RemediationError{
		Inner:       fmt.Errorf(msg),
		Remediation: fsterr.TokenExpirationRemediation,
	}
}
