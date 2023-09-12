package authenticate

import (
	"fmt"
	"io"
	"time"

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
	authServer  auth.Starter
	openBrowser func(string) error

	// IMPORTANT: The following fields are public to the `profile` subcommands.

	// NewProfile indicates if we should create a new profile.
	NewProfile bool
	// NewProfileName indicates the new profile name.
	NewProfileName string
	// ProfileDefault indicates if the affected profile should become the default.
	ProfileDefault bool
	// UpdateProfile indicates if we should update a profile.
	UpdateProfile bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data, opener func(string) error, authServer auth.Starter) *RootCommand {
	var c RootCommand
	c.authServer = authServer
	c.openBrowser = opener
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

	accountEndpoint, _ := c.Globals.Account()
	apiEndpoint, _ := c.Globals.Endpoint()
	verifier, err := auth.GenVerifier()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate a code verifier: %w", err),
			Remediation: auth.Remediation,
		}
	}
	c.authServer.SetAccountEndpoint(accountEndpoint)
	c.authServer.SetAPIEndpoint(apiEndpoint)
	c.authServer.SetVerifier(verifier)

	var serverErr error
	go func() {
		err := c.authServer.Start()
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

	err = c.openBrowser(authorizationURL)
	if err != nil {
		return fmt.Errorf("failed to open your default browser: %w", err)
	}

	ar := <-c.authServer.GetResult()
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

// processProfiles updates the relevant profile with the returned token data.
//
// First it checks the --profile flag and the `profile` fastly.toml field.
// Second it checks to see which profile is currently the default.
// Third it identifies which profile to be modified.
// Fourth it writes the updated in-memory data back to disk.
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
	case noProfilesConfigured(profileOverride, profileDefault):
		makeDefault := true
		c.Globals.Config.Profiles = createNewProfile(profile.DefaultName, makeDefault, c.Globals.Config.Profiles, ar)
	case invokedByProfileCreateCommand(c):
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
		makeDefault := c.ProfileDefault // this is set by `profile update` command.
		if !c.UpdateProfile {           // if not invoked by `profile update`, then get current `Default` field value
			if p := profile.Get(profileName, c.Globals.Config.Profiles); p != nil {
				makeDefault = p.Default
			}
		}
		ps, err := editProfile(profileName, makeDefault, c.Globals.Config.Profiles, ar)
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

// noProfilesConfigured determines if no profiles have been defined.
func noProfilesConfigured(o, d string) bool {
	return o == "" && d == ""
}

// invokedByProfileCreateCommand determines if this command was invoked by the
// `profile create` subcommand.
func invokedByProfileCreateCommand(c *RootCommand) bool {
	return c.NewProfile && c.NewProfileName != ""
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
func editProfile(profileName string, makeDefault bool, p config.Profiles, ar auth.AuthorizationResult) (config.Profiles, error) {
	ps, ok := profile.Edit(profileName, p, func(p *config.Profile) {
		now := time.Now().Unix()
		p.Default = makeDefault
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
