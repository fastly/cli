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

	// IMPORTANT: The following fields are public to the `profile` subcommands.

	// NewProfile indicates if we should create a new profile.
	NewProfile bool
	// NewProfileName indicates the new profile name.
	NewProfileName string
	// ProfileDefault indicates if the affected profile should become the default.
	ProfileDefault bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("authenticate", "SSO (Single Sign-On) authentication")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	// We need to prompt the user, so they know we're about to open their web
	// browser, but we also need to handle the scenario where the `authenticate`
	// command is invoked indirectly via ../../app/run.go as that package will
	// have its own (similar) prompt before invoking this command. So to avoid a
	// double prompt, the app package will set `SkipAuthPrompt: true`.
	if !c.Globals.SkipAuthPrompt && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Warning(out, "We need to open your browser to authenticate you.")
		text.Break(out)
		cont, err := text.AskYesNo(out, text.BoldYellow("Do you want to continue? [y/N]: "), in)
		text.Break(out)
		if err != nil {
			return err
		}
		if !cont {
			return fmt.Errorf("user cancelled execution")
		}
	}

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

	err = c.processProfiles(ar)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to process profile data: %w", err)
	}

	text.Success(out, "Session token (persisted to your local configuration): %s", ar.SessionToken)
	return nil
}

func (c *RootCommand) processProfiles(ar auth.AuthorizationResult) error {
	var profileOverride string
	switch {
	case c.Globals.Flags.Profile != "":
		profileOverride = c.Globals.Flags.Profile
	case c.Globals.Manifest.File.Profile != "":
		profileOverride = c.Globals.Manifest.File.Profile
	}

	profileDefault, _ := profile.Default(c.Globals.Config.Profiles)

	switch {
	case profileOverride == "" && profileDefault == "": // If no profiles configured at all, create a new default...
		makeDefault := true
		c.Globals.Config.Profiles = createNewProfile(profile.DefaultName, makeDefault, c.Globals.Config.Profiles, ar)
	case c.NewProfile: // We know we've been triggered by `profile create` if this is set.
		c.Globals.Config.Profiles = createNewProfile(c.NewProfileName, c.ProfileDefault, c.Globals.Config.Profiles, ar)

		// If the user wants the newly created profile to be their new default, then
		// we'll call Set for its side effect of resetting all other profiles to have
		// their Default field set to false.
		if c.ProfileDefault {
			if p, ok := profile.SetDefault(c.NewProfileName, c.Globals.Config.Profiles); ok {
				c.Globals.Config.Profiles = p
			}
		}
	default:
		// Otherwise, edit the existing profile to have the newly generated tokens.
		profileName := profileDefault
		if profileOverride != "" {
			profileName = profileOverride
		}
		ps, err := editProfile(profileName, c.ProfileDefault, c.Globals.Config.Profiles, ar)
		if err != nil {
			return err
		}
		c.Globals.Config.Profiles = ps
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}
	return nil
}

// IMPORTANT: Mutates the config.Profiles map type.
// We need to return the modified type so it can be safely reassigned.
func createNewProfile(profileName string, makeDefault bool, p config.Profiles, ar auth.AuthorizationResult) config.Profiles {
	now := time.Now().Unix()
	if p == nil {
		p = make(config.Profiles)
	}
	p[profileName] = &config.Profile{
		AccessToken:         ar.Jwt.AccessToken,
		AccessTokenCreated:  now,
		AccessTokenTTL:      ar.Jwt.ExpiresIn,
		Default:             makeDefault,
		Email:               ar.Email,
		RefreshToken:        ar.Jwt.RefreshToken,
		RefreshTokenCreated: now,
		RefreshTokenTTL:     ar.Jwt.RefreshExpiresIn,
		Token:               ar.SessionToken,
	}
	return p
}

// IMPORTANT: Mutates the config.Profiles map type.
// We need to return the modified type so it can be safely reassigned.
func editProfile(profileName string, profileDefault bool, p config.Profiles, ar auth.AuthorizationResult) (config.Profiles, error) {
	ps, ok := profile.Edit(profileName, p, func(p *config.Profile) {
		now := time.Now().Unix()
		p.Default = profileDefault
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
		return ps, fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to update '%s' profile with new token data", profileName),
			Remediation: "Run `fastly authenticate` to retry.",
		}
	}
	return ps, nil
}
