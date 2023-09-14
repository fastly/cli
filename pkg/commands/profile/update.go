package profile

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/sso"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// APIClientFactory allows the profile command to regenerate the global Fastly
// API client when a new token is provided, in order to validate that token.
// It's a redeclaration of the app.APIClientFactory to avoid an import loop.
type APIClientFactory func(token, endpoint string, debugMode bool) (api.Interface, error)

// UpdateCommand represents a Kingpin command.
type UpdateCommand struct {
	cmd.Base
	authCmd *sso.RootCommand

	automationToken bool
	clientFactory   APIClientFactory
	profile         string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, cf APIClientFactory, g *global.Data, authCmd *sso.RootCommand) *UpdateCommand {
	var c UpdateCommand
	c.Globals = g
	c.authCmd = authCmd
	c.CmdClause = parent.Command("update", "Update user profile")
	c.CmdClause.Arg("profile", "Profile to update (defaults to the currently active profile)").Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("automation-token", "Expected input will be an 'automation token' instead of a 'user token'").BoolVar(&c.automationToken)
	c.clientFactory = cf
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	profileName, p, err := c.identifyProfile()
	if err != nil {
		return fmt.Errorf("failed to identify the profile to update: %w", err)
	}
	text.Info(out, "Profile being updated: '%s'.\n\n", profileName)

	// Set to true for --auto-yes/--non-interactive flags, otherwise prompt user.
	makeDefault := true
	updateToken := true

	if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		makeDefault, err = text.AskYesNo(out, text.BoldYellow("Make profile the default? [y/N] "), in)
		text.Break(out)
		if err != nil {
			return err
		}

		updateToken, err = text.AskYesNo(out, text.BoldYellow("Update the token associated with this profile? [y/N]: "), in)
		if err != nil {
			return err
		}
	}

	if updateToken {
		err := c.updateToken(profileName, makeDefault, p, in, out)
		if err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}
	} else if makeDefault { // only set default if not updating the token (as updating the token already handles setting the default)
		err := c.updateDefault(profileName)
		if err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}
	}

	text.Success(out, "\nProfile '%s' updated", profileName)
	return nil
}

func (c *UpdateCommand) identifyProfile() (string, *config.Profile, error) {
	var (
		profileName string
		p           *config.Profile
	)

	// If profile argument not set and no --profile flag set, then identify the
	// default profile to update.
	if c.profile == "" && c.Globals.Flags.Profile == "" {
		profileName, p = profile.Default(c.Globals.Config.Profiles)
		if p == nil {
			return "", nil, fsterr.RemediationError{
				Inner:       fmt.Errorf("no active profile"),
				Remediation: profile.NoDefaults,
			}
		}
	} else {
		// Otherwise, acquire the profile the user has specified.
		profileName = c.profile
		if c.Globals.Flags.Profile != "" {
			profileName = c.Globals.Flags.Profile
		}
		p = profile.Get(profileName, c.Globals.Config.Profiles)
		if p == nil {
			msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
			return "", nil, fsterr.RemediationError{
				Inner:       fmt.Errorf(msg),
				Remediation: fsterr.ProfileRemediation,
			}
		}
	}

	return profileName, p, nil
}

func (c *UpdateCommand) updateToken(profileName string, makeDefault bool, p *config.Profile, in io.Reader, out io.Writer) error {
	// Prompt user to decide on which token flow to take (OAuth or Static)
	// Static being a traditional long-lived user/automation token.
	//
	// Opting for the OAuth flow will generate a short-lived token.
	// Otherwise user has to create a token manually, then paste it when prompted.
	useOAuthFlow := true
	if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Info(out, "When updating a profile you can either paste in a long-lived token or allow the Fastly CLI to regenerate a short-lived token that can be automatically refreshed.")
		text.Break(out)
		var err error
		useOAuthFlow, err = text.AskYesNo(out, text.BoldYellow("Continue with Fastly SSO (Single Sign-On) authentication for generating a short-lived token? [y/N]: "), in)
		text.Break(out)
		if err != nil {
			return err
		}
	}

	if useOAuthFlow {
		// IMPORTANT: We need to set profile fields for `sso` command.
		//
		// This is so the `sso` command will use this information to update
		// the specific profile.
		c.authCmd.InvokedFromProfileUpdate = true
		c.authCmd.ProfileUpdateName = profileName
		c.authCmd.ProfileDefault = makeDefault

		// NOTE: The `sso` command already handles writing config back to disk.
		// So unlike `c.staticTokenFlow` (below) we don't have to do that here.
		err := c.authCmd.Exec(in, out)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		text.Break(out)
	} else {
		if err := c.staticTokenFlow(profileName, makeDefault, p, in, out); err != nil {
			return fmt.Errorf("failed to process the static token flow: %w", err)
		}
		// Write the in-memory representation back to disk.
		if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error saving config file: %w", err)
		}
	}

	return nil
}

func (c *UpdateCommand) updateDefault(profileName string) error {
	p, ok := profile.SetDefault(profileName, c.Globals.Config.Profiles)
	if !ok {
		return errors.New("failed to update the profile's default field")
	}
	c.Globals.Config.Profiles = p

	// Write the in-memory representation back to disk.
	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}
	return nil
}

// validateToken ensures the token can be used to acquire user data.
func (c *UpdateCommand) validateToken(token, endpoint string, spinner text.Spinner) (string, error) {
	var (
		client api.Interface
		err    error
		t      *fastly.Token
	)
	err = spinner.Process("Validating token", func(_ *text.SpinnerWrapper) error {
		client, err = c.clientFactory(token, endpoint, c.Globals.Flags.Debug)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Endpoint": endpoint,
			})
			return fmt.Errorf("error regenerating Fastly API client: %w", err)
		}

		t, err = client.GetTokenSelf()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error validating token: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if c.automationToken {
		return fmt.Sprintf("Automation Token (%s)", t.ID), nil
	}

	var user *fastly.User
	err = spinner.Process("Getting user data", func(_ *text.SpinnerWrapper) error {
		user, err = client.GetUser(&fastly.GetUserInput{
			ID: t.UserID,
		})
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"User ID": t.UserID,
			})
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("error fetching token user: %w", err),
				Remediation: "If providing an 'automation token', retry the command with the `--automation-token` flag set.",
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return user.Login, nil
}

func (c *UpdateCommand) staticTokenFlow(profileName string, makeDefault bool, p *config.Profile, in io.Reader, out io.Writer) error {
	opts := []profile.EditOption{}

	token, err := text.InputSecure(out, text.BoldYellow("Profile token: (leave blank to skip): "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	// User didn't want to change their token value so reassign original.
	if token == "" {
		token = p.Token
	} else {
		opts = append(opts, func(p *config.Profile) {
			p.Token = token
		})
	}
	text.Break(out)

	opts = append(opts, func(p *config.Profile) {
		p.Default = makeDefault
	})

	text.Break(out)

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
		}
	}()

	endpoint, _ := c.Globals.Endpoint()

	email, err := c.validateToken(token, endpoint, spinner)
	if err != nil {
		return err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Email = email
	})

	ps, ok := profile.Edit(profileName, c.Globals.Config.Profiles, opts...)
	if !ok {
		msg := fmt.Sprintf(profile.DoesNotExist, profileName)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}
	c.Globals.Config.Profiles = ps

	if makeDefault {
		// We call SetDefault for its side effect of resetting all other profiles to have
		// their Default field set to false.
		ps, ok = profile.SetDefault(profileName, ps)
		if !ok {
			msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
			err := errors.New(msg)
			c.Globals.ErrLog.Add(err)
			return err
		}
	}
	c.Globals.Config.Profiles = ps

	return nil
}
