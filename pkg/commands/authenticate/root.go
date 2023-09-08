package authenticate

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/cap/oidc"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("authenticate", "Authenticate with Fastly (returns temporary, auto-rotated, API token)")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	verifier, err := oidc.NewCodeVerifier()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate a code verifier: %w", err),
			Remediation: auth.Remediation,
		}
	}

	result := make(chan auth.AuthorizationResult)
	apiEndpoint, _ := c.Globals.Endpoint()
	accountEndpoint, _ := c.Globals.Account()

	s := auth.Server{
		APIEndpoint:     apiEndpoint,
		AccountEndpoint: accountEndpoint,
		HTTPClient:      c.Globals.HTTPClient,
		Result:          result,
		Router:          http.NewServeMux(),
		Verifier:        verifier,
	}
	s.Routes()

	var serverErr error

	go func() {
		err := s.Start()
		if err != nil {
			serverErr = err
		}
	}()

	if serverErr != nil {
		return serverErr
	}

	text.Info(out, "Starting a local server to handle the authentication flow.")

	authorizationURL, err := auth.GenURL(accountEndpoint, apiEndpoint, verifier)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate an authorization URL: %w", err),
			Remediation: auth.Remediation,
		}
	}

	text.Break(out)
	text.Description(out, "We're opening the following URL in your default web browser so you may authenticate with Fastly", authorizationURL)

	err = open.Run(authorizationURL)
	if err != nil {
		return fmt.Errorf("failed to open your default browser: %w", err)
	}

	ar := <-result
	if ar.Err != nil || ar.SessionToken == "" {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to authorize: %w", ar.Err),
			Remediation: auth.Remediation,
		}
	}

	var profileConfigured string
	switch {
	case c.Globals.Flags.Profile != "":
		profileConfigured = c.Globals.Flags.Profile
	case c.Globals.Manifest.File.Profile != "":
		profileConfigured = c.Globals.Manifest.File.Profile
	}

	profileDefault, _ := profile.Default(c.Globals.Config.Profiles)

	// If no profiles configured at all, create a new default...
	if profileConfigured == "" && profileDefault == "" {
		now := time.Now().Unix()
		if c.Globals.Config.Profiles == nil {
			c.Globals.Config.Profiles = make(config.Profiles)
		}
		c.Globals.Config.Profiles[profile.DefaultName] = &config.Profile{
			AccessToken:         ar.Jwt.AccessToken,
			AccessTokenCreated:  now,
			AccessTokenTTL:      ar.Jwt.ExpiresIn,
			Default:             true,
			Email:               ar.Email,
			RefreshToken:        ar.Jwt.RefreshToken,
			RefreshTokenCreated: now,
			RefreshTokenTTL:     ar.Jwt.RefreshExpiresIn,
			Token:               ar.SessionToken,
		}
	} else {
		// Otherwise, edit the default to have the newly generated tokens.
		profileName := profileDefault
		if profileConfigured != "" {
			profileName = profileConfigured
		}
		ps, ok := profile.Edit(profileName, c.Globals.Config.Profiles, func(p *config.Profile) {
			now := time.Now().Unix()
			p.AccessToken = ar.Jwt.AccessToken
			p.AccessTokenCreated = now
			p.AccessTokenTTL = ar.Jwt.ExpiresIn
			p.Email = ar.Email
			p.RefreshToken = ar.Jwt.RefreshToken
			p.RefreshTokenCreated = now
			p.RefreshTokenTTL = ar.Jwt.RefreshExpiresIn
			p.Token = ar.SessionToken
		})
		if !ok {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to update default profile with new token data"),
				Remediation: "Run `fastly authenticate` to retry.",
			}
		}
		c.Globals.Config.Profiles = ps
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	text.Success(out, "Session token (persisted to your local configuration): %s", ar.SessionToken)
	return nil
}
