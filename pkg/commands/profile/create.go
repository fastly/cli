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
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand represents a Kingpin command.
type CreateCommand struct {
	cmd.Base

	automationToken bool
	clientFactory   APIClientFactory
	profile         string
}

// NewCreateCommand returns a new command registered in the parent.
func NewCreateCommand(parent cmd.Registerer, cf APIClientFactory, g *global.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = g
	c.CmdClause = parent.Command("create", "Create user profile")
	c.CmdClause.Arg("profile", "Profile to create (default 'user')").Default("user").Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("automation-token", "Expected input will be an 'automation token' instead of a 'user token'").BoolVar(&c.automationToken)
	c.clientFactory = cf
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if profile.Exist(c.profile, c.Globals.Config.Profiles) {
		return fmt.Errorf("profile '%s' already exists", c.profile)
	}

	// The Default status of a new profile should always be true unless there is
	// an existing profile already set to be the default. In the latter scenario
	// we should prompt the user to see if the new profile they're creating needs
	// to become the new default.
	def := true
	if profileName, _ := profile.Default(c.Globals.Config.Profiles); profileName != "" {
		def, err = c.promptForDefault(in, out)
		if err != nil {
			return err
		}
	}

	if err := c.tokenFlow(def, in, out); err != nil {
		return err
	}
	if err := c.persistCfg(); err != nil {
		return err
	}

	displayCfgPath(c.Globals.ConfigPath, out)
	text.Success(out, "Profile '%s' created", c.profile)
	return nil
}

// tokenFlow initialises the token flow.
func (c *CreateCommand) tokenFlow(def bool, in io.Reader, out io.Writer) error {
	token, err := promptForToken(in, out, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	text.Break(out)

	endpoint, _ := c.Globals.Endpoint()

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

	return c.updateInMemCfg(email, token, endpoint, def, spinner)
}

func promptForToken(in io.Reader, out io.Writer, errLog fsterr.LogInterface) (string, error) {
	text.Output(out, "\nAn API token is used to authenticate requests to the Fastly API. To create a token, visit https://manage.fastly.com/account/personal/tokens\n\n")
	token, err := text.InputSecure(out, text.BoldYellow("Fastly API token: "), in, validateTokenNotEmpty)
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
		client, err = c.clientFactory(token, endpoint)
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
func (c *CreateCommand) updateInMemCfg(email, token, endpoint string, def bool, spinner text.Spinner) error {
	return spinner.Process("Persisting configuration", func(_ *text.SpinnerWrapper) error {
		c.Globals.Config.Fastly.APIEndpoint = endpoint

		if c.Globals.Config.Profiles == nil {
			c.Globals.Config.Profiles = make(config.Profiles)
		}
		c.Globals.Config.Profiles[c.profile] = &config.Profile{
			Default: def,
			Email:   email,
			Token:   token,
		}

		// If the user wants the newly created profile to be their new default, then
		// we'll call SetDefault for its side effect of resetting all other profiles
		// to have their Default field set to false.
		if def {
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
	cont, err := text.AskYesNo(out, text.BoldYellow("Set this profile to be your default? [y/N] "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return false, err
	}
	return cont, nil
}
