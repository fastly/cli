package profile

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/authenticate"
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
	authCmd *authenticate.RootCommand

	automationToken bool
	clientFactory   APIClientFactory
	profile         string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, cf APIClientFactory, g *global.Data, authCmd *authenticate.RootCommand) *UpdateCommand {
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
	var (
		profileName string
		p           *config.Profile
	)

	if c.profile == "" && c.Globals.Flags.Profile == "" {
		profileName, p = profile.Default(c.Globals.Config.Profiles)
		if profileName == "" {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("no active profile"),
				Remediation: profile.NoDefaults,
			}
		}
	} else {
		profileName = c.profile
		if c.Globals.Flags.Profile != "" {
			profileName = c.Globals.Flags.Profile
		}
		profileName, p = profile.Get(profileName, c.Globals.Config.Profiles)
		if profileName == "" {
			msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
			return fsterr.RemediationError{
				Inner:       fmt.Errorf(msg),
				Remediation: fsterr.ProfileRemediation,
			}
		}
	}

	text.Info(out, "Profile being updated: '%s'.\n\n", profileName)

	makeDefault, opts, err := c.staticTokenFlow(p, in, out)
	if err != nil {
		return fmt.Errorf("failed to process the static token flow: %w", err)
	}

	var ok bool
	ps, ok := profile.Edit(profileName, c.Globals.Config.Profiles, opts...)
	if !ok {
		msg := fmt.Sprintf(profile.DoesNotExist, profileName)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	if makeDefault {
		// We call SetDefault for its side effect of resetting all other profiles to have
		// their Default field set to false.
		ps, ok = profile.SetDefault(c.profile, ps)
		if !ok {
			msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
			err := errors.New(msg)
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	c.Globals.Config.Profiles = ps

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	text.Success(out, "\nProfile '%s' updated", profileName)
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

func (c *UpdateCommand) staticTokenFlow(p *config.Profile, in io.Reader, out io.Writer) (bool, []profile.EditOption, error) {
	var makeDefault bool
	opts := []profile.EditOption{}

	token, err := text.InputSecure(out, text.BoldYellow("Profile token: (leave blank to skip): "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return makeDefault, opts, err
	}
	if token != "" {
		opts = append(opts, func(p *config.Profile) {
			p.Token = token
		})
	}

	text.Break(out)
	text.Break(out)

	makeDefault, err = text.AskYesNo(out, "Make profile the default? [y/N] ", in)
	if err != nil {
		return makeDefault, opts, err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Default = makeDefault
	})

	// User didn't want to change their token value so reassign original.
	if token == "" {
		token = p.Token
	}

	text.Break(out)

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return makeDefault, opts, err
	}
	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
		}
	}()

	endpoint, _ := c.Globals.Endpoint()

	email, err := c.validateToken(token, endpoint, spinner)
	if err != nil {
		return makeDefault, opts, err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Email = email
	})

	return makeDefault, opts, nil
}
