package profile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/sso"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand represents a Kingpin command.
type CreateCommand struct {
	argparser.Base
	authCmd *sso.RootCommand

	automationToken bool
	profile         string
	sso             bool
}

// NewCreateCommand returns a new command registered in the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data, authCmd *sso.RootCommand) *CreateCommand {
	var c CreateCommand
	c.Globals = g
	c.authCmd = authCmd
	c.CmdClause = parent.Command("create", "Create user profile")
	c.CmdClause.Arg("profile", "Profile to create (default 'user')").Default(profile.DefaultName).Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("automation-token", "Expected input will be an 'automation token' instead of a 'user token'").BoolVar(&c.automationToken)
	c.CmdClause.Flag("sso", "Create an SSO-based token").Hidden().BoolVar(&c.sso)
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.sso && c.automationToken {
		return fsterr.ErrInvalidProfileSSOCombo
	}

	if profile.Exist(c.profile, c.Globals.Config.Profiles) {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("profile '%s' already exists", c.profile),
			Remediation: "Re-run the command and pass a different value for the 'profile' argument.",
		}
	}

	// FIXME: Put back messaging once SSO is GA.
	// if !c.sso {
	// 	text.Info(out, "When creating a profile you can either paste in a long-lived token or allow the Fastly CLI to generate a short-lived token that can be automatically refreshed. To create an SSO-based token, pass the `--sso` flag: `fastly profile create --sso`.")
	// 	text.Break(out)
	// }

	// The Default status of a new profile should always be true unless there is
	// an existing profile already set to be the default. In the latter scenario
	// we should prompt the user to see if the new profile they're creating needs
	// to become the new default.
	makeDefault := true
	if _, p := profile.Default(c.Globals.Config.Profiles); p != nil && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		makeDefault, err = c.promptForDefault(in, out)
		if err != nil {
			return err
		}
	}

	if c.sso {
		// IMPORTANT: We need to set profile fields for `sso` command.
		//
		// This is so the `sso` command will use this information to create
		// a new 'non-default' profile.
		c.authCmd.InvokedFromProfileCreate = true
		c.authCmd.ProfileCreateName = c.profile
		c.authCmd.ProfileDefault = makeDefault

		err = c.authCmd.Exec(in, out)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		text.Break(out)
	} else {
		if err := c.staticTokenFlow(makeDefault, in, out); err != nil {
			return err
		}
	}

	if err := c.persistCfg(); err != nil {
		return err
	}

	displayCfgPath(c.Globals.ConfigPath, out)
	text.Success(out, "Profile '%s' created", c.profile)
	return nil
}

// staticTokenFlow initialises the token flow for a non-OAuth token.
func (c *CreateCommand) staticTokenFlow(makeDefault bool, in io.Reader, out io.Writer) error {
	token, err := promptForToken(in, out, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	text.Break(out)

	endpoint, _ := c.Globals.APIEndpoint()

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
		}
	}()

	email, err := c.validateToken(token, endpoint, spinner)
	if err != nil {
		return err
	}

	return c.updateInMemCfg(email, token, endpoint, makeDefault, spinner)
}

func promptForToken(in io.Reader, out io.Writer, errLog fsterr.LogInterface) (string, error) {
	text.Output(out, "\nAn API token is used to authenticate requests to the Fastly API. To create a token, visit https://manage.fastly.com/account/personal/tokens\n\n")
	token, err := text.InputSecure(out, text.Prompt("Fastly API token: "), in, validateTokenNotEmpty)
	if err != nil {
		errLog.Add(err)
		return "", err
	}
	text.Break(out)
	return token, nil
}

func validateTokenNotEmpty(s string) error {
	if s == "" {
		return ErrEmptyToken
	}
	return nil
}

// ErrEmptyToken is returned when a user tries to supply an empty string as a
// token in the terminal prompt.
var ErrEmptyToken = errors.New("token cannot be empty")

// validateToken ensures the token can be used to acquire user data.
func (c *CreateCommand) validateToken(token, endpoint string, spinner text.Spinner) (string, error) {
	var (
		client api.Interface
		err    error
		t      *fastly.Token
	)
	err = spinner.Process("Validating token", func(_ *text.SpinnerWrapper) error {
		client, err = c.Globals.APIClientFactory(token, endpoint, c.Globals.Flags.Debug)
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

// updateInMemCfg persists the updated configuration data in-memory.
func (c *CreateCommand) updateInMemCfg(email, token, endpoint string, makeDefault bool, spinner text.Spinner) error {
	return spinner.Process("Persisting configuration", func(_ *text.SpinnerWrapper) error {
		c.Globals.Config.Fastly.APIEndpoint = endpoint

		if c.Globals.Config.Profiles == nil {
			c.Globals.Config.Profiles = make(config.Profiles)
		}
		c.Globals.Config.Profiles[c.profile] = &config.Profile{
			Default: makeDefault,
			Email:   email,
			Token:   token,
		}

		// If the user wants the newly created profile to be their new default, then
		// we'll call SetDefault for its side effect of resetting all other profiles
		// to have their Default field set to false.
		if makeDefault {
			if p, ok := profile.SetDefault(c.profile, c.Globals.Config.Profiles); ok {
				c.Globals.Config.Profiles = p
			}
		}
		return nil
	})
}

func (c *CreateCommand) persistCfg() error {
	// TODO: The following directory checks should be encapsulated by the
	// File.Write() method as this chunk of code is duplicated in various places.
	// Consider consolidating with pkg/filesystem/directory.go
	// This function is itself duplicated in pkg/commands/profile/update.go
	dir := filepath.Dir(c.Globals.ConfigPath)
	fi, err := os.Stat(dir)
	switch {
	case err == nil && !fi.IsDir():
		return fmt.Errorf("config file path %s isn't a directory", dir)
	case err != nil && errors.Is(err, fs.ErrNotExist):
		if err := os.MkdirAll(dir, config.DirectoryPermissions); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Directory":   dir,
				"Permissions": config.DirectoryPermissions,
			})
			return fmt.Errorf("error creating config file directory: %w", err)
		}
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	return nil
}

func displayCfgPath(path string, out io.Writer) {
	filePath := strings.ReplaceAll(path, " ", `\ `)
	text.Break(out)
	text.Description(out, "You can find your configuration file at", filePath)
}

func (c *CreateCommand) promptForDefault(in io.Reader, out io.Writer) (bool, error) {
	cont, err := text.AskYesNo(out, "Set this profile to be your default? [y/N] ", in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return false, err
	}
	return cont, nil
}
