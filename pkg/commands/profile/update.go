package profile

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// APIClientFactory allows the profile command to regenerate the global Fastly
// API client when a new token is provided, in order to validate that token.
// It's a redeclaration of the app.APIClientFactory to avoid an import loop.
type APIClientFactory func(token, endpoint string) (api.Interface, error)

// UpdateCommand represents a Kingpin command.
type UpdateCommand struct {
	cmd.Base

	automationToken bool
	clientFactory   APIClientFactory
	profile         string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, cf APIClientFactory, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = g
	c.CmdClause = parent.Command("update", "Update user profile")
	c.CmdClause.Arg("profile", "Profile to update (defaults to the currently active profile)").Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("automation-token", "Expected input will be an 'automation token' instead of a 'user token'").BoolVar(&c.automationToken)
	c.clientFactory = cf
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	var (
		name string
		p    *config.Profile
	)

	if c.profile == "" {
		name, p = profile.Default(c.Globals.Config.Profiles)
		if name == "" {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("no active profile"),
				Remediation: profile.NoDefaults,
			}
		}
	} else {
		name, p = profile.Get(c.profile, c.Globals.Config.Profiles)
		if name == "" {
			msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
			return fsterr.RemediationError{
				Inner:       fmt.Errorf(msg),
				Remediation: fsterr.ProfileRemediation,
			}
		}
	}

	text.Info(out, "Profile being updated: '%s'.", name)

	opts := []profile.EditOption{}

	text.Break(out)

	token, err := text.InputSecure(out, text.BoldYellow("Profile token: (leave blank to skip): "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if token != "" {
		opts = append(opts, func(p *config.Profile) {
			p.Token = token
		})
	}

	text.Break(out)
	text.Break(out)

	def, err := text.AskYesNo(out, "Make profile the default? [y/N] ", in)
	if err != nil {
		return err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Default = def
	})

	// User didn't want to change their token value so reassign original.
	if token == "" {
		token = p.Token
	}

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

	var ok bool

	ps, ok := profile.Edit(name, c.Globals.Config.Profiles, opts...)
	if !ok {
		msg := fmt.Sprintf(profile.DoesNotExist, name)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}
	c.Globals.Config.Profiles = ps

	if err := c.Globals.Config.Write(c.Globals.Path); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	text.Success(out, "Profile '%s' updated", name)
	return nil
}

// validateToken ensures the token can be used to acquire user data.
func (c *UpdateCommand) validateToken(token, endpoint string, spinner text.Spinner) (string, error) {
	err := spinner.Start()
	if err != nil {
		return "", err
	}
	msg := "Validating token"
	spinner.Message(msg + "...")

	client, err := c.clientFactory(token, endpoint)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			fsterr.AllowInstrumentation: true,
			"Endpoint":                  endpoint,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return "", spinErr
		}

		return "", fmt.Errorf("error regenerating Fastly API client: %w", err)
	}

	t, err := client.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return "", spinErr
		}

		return "", fmt.Errorf("error validating token: %w", err)
	}

	if c.automationToken {
		spinner.StopMessage(msg)
		err = spinner.Stop()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Automation Token (%s)", t.ID), nil
	}

	user, err := client.GetUser(&fastly.GetUserInput{
		ID: t.UserID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User ID": t.UserID,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return "", spinErr
		}

		return "", fmt.Errorf("error fetching token user: %w", err)
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return "", err
	}
	return user.Login, nil
}
