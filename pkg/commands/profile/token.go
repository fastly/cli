package profile

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	fsttime "github.com/fastly/cli/pkg/time"
)

// TokenCommand represents a Kingpin command.
type TokenCommand struct {
	argparser.Base
	Profile  string        `arg:"" optional:"" help:"Print API token for the named profile"`
	TokenTTL time.Duration `name:"ttl" default:"5m0s" help:"Amount of time for which the token must be valid (in seconds 's', minutes 'm', or hours 'h')"`
}

const defaultTokenTTL time.Duration = 5 * time.Minute

// Exec implements the command interface.
func (c *TokenCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	if !c.Globals.Flags.Quiet {
		text.Deprecated("This command will be removed in a future release. Use 'fastly auth show' instead.\n\n")
	}

	var name string
	if c.Profile != "" {
		name = c.Profile
	}
	if c.Globals.Flags.Profile != "" {
		name = c.Globals.Flags.Profile
	}

	if name != "" {
		at := c.Globals.Config.GetAuthToken(name)
		if at != nil {
			if err = checkTokenValidity(name, at, c.TokenTTL); err != nil {
				return err
			}
			text.Output(out, at.Token)
			return nil
		}
		msg := fmt.Sprintf("the profile '%s' does not exist", name)
		return fsterr.RemediationError{
			Inner:       errors.New(msg),
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	if name, at := c.Globals.Config.GetDefaultAuthToken(); at != nil {
		if err = checkTokenValidity(name, at, c.TokenTTL); err != nil {
			return err
		}
		text.Output(out, at.Token)
		return nil
	}
	return fsterr.RemediationError{
		Inner:       errors.New("no profiles available"),
		Remediation: fsterr.ProfileRemediation(),
	}
}

func checkTokenValidity(name string, at *config.AuthToken, ttl time.Duration) error {
	var expiryStr string
	if at.Type == config.AuthTokenTypeSSO {
		expiryStr = at.RefreshExpiresAt
	} else {
		expiryStr = at.APITokenExpiresAt
	}
	if expiryStr == "" {
		return nil
	}

	expiry, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return err
	}

	if expiry.After(time.Now().Add(ttl)) {
		return nil
	}

	var msg string
	if expiry.Before(time.Now()) {
		msg = fmt.Sprintf("the token in profile '%s' expired at '%s'", name, expiry.UTC().Format(fsttime.Format))
	} else {
		msg = fmt.Sprintf("the token in profile '%s' will expire at '%s'", name, expiry.UTC().Format(fsttime.Format))
	}

	return fsterr.RemediationError{
		Inner:       errors.New(msg),
		Remediation: fsterr.TokenExpirationRemediation(),
	}
}
