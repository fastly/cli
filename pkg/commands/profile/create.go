package profile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand represents a Kingpin command.
type CreateCommand struct {
	cmd.Base

	clientFactory APIClientFactory
	profile       string
}

// NewCreateCommand returns a new command registered in the parent.
func NewCreateCommand(parent cmd.Registerer, cf APIClientFactory, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("create", "Create user profile")
	c.CmdClause.Arg("profile", "Profile to create (default 'user')").Default("user").Short('p').StringVar(&c.profile)
	c.clientFactory = cf
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if profile.Exist(c.profile, c.Globals.File.Profiles) {
		return fmt.Errorf("profile '%s' already exists", c.profile)
	}

	// The Default status of a new profile should always be true unless there is
	// an existing profile already set to be the default. In the latter scenario
	// we should prompt the user to see if the new profile they're creating needs
	// to become the new default.
	def := true
	if profile, _ := profile.Default(c.Globals.File.Profiles); profile != "" {
		def, err = c.promptForDefault(in, out)
		if err != nil {
			return err
		}
	}

	if err := c.tokenFlow(c.profile, def, in, out); err != nil {
		return err
	}
	if err := c.persistCfg(); err != nil {
		return err
	}

	displayCfgPath(c.Globals.Path, out)
	text.Success(out, "Profile '%s' created", c.profile)
	return nil
}

// tokenFlow initialises the token flow.
func (c *CreateCommand) tokenFlow(profile string, def bool, in io.Reader, out io.Writer) error {
	var err error

	// If user provides a --token flag, then don't prompt them for input.
	token, source := c.Globals.Token()
	if source == config.SourceFile || source == config.SourceUndefined {
		token, err = promptForToken(in, out, c.Globals.ErrLog)
		if err != nil {
			return err
		}
		text.Break(out)
		text.Break(out)
	}

	progress := text.NewProgress(out, c.Globals.Verbose())
	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
	}()

	endpoint, _ := c.Globals.Endpoint()

	user, err := c.validateToken(token, endpoint, progress)
	if err != nil {
		return err
	}

	c.updateInMemCfg(profile, user.Login, token, endpoint, def, progress)

	progress.Done()
	return nil
}

func promptForToken(in io.Reader, out io.Writer, errLog fsterr.LogInterface) (string, error) {
	text.Break(out)
	text.Output(out, `An API token is used to authenticate requests to the Fastly API.
			To create a token, visit https://manage.fastly.com/account/personal/tokens
		`)
	text.Break(out)
	token, err := text.InputSecure(out, "Fastly API token: ", in, validateTokenNotEmpty)
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

// ErrEmptyToken is returned when a user tries to supply an emtpy string as a
// token in the terminal prompt.
var ErrEmptyToken = errors.New("token cannot be empty")

// validateToken ensures the token can be used to acquire user data.
func (c *CreateCommand) validateToken(token, endpoint string, progress text.Progress) (*fastly.User, error) {
	progress.Step("Validating token...")

	client, err := c.clientFactory(token, endpoint)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Endpoint": endpoint,
		})
		return nil, fmt.Errorf("error regenerating Fastly API client: %w", err)
	}

	t, err := client.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	user, err := client.GetUser(&fastly.GetUserInput{
		ID: t.UserID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"User ID": t.UserID,
		})
		return nil, fmt.Errorf("error fetching token user: %w", err)
	}

	return user, nil
}

// updateInMemCfg persists the updated configuration data in-memory.
func (c *CreateCommand) updateInMemCfg(profileName, email, token, endpoint string, def bool, progress text.Progress) {
	progress.Step("Persisting configuration...")

	c.Globals.File.Fastly.APIEndpoint = endpoint

	if c.Globals.File.Profiles == nil {
		c.Globals.File.Profiles = make(config.Profiles)
	}
	c.Globals.File.Profiles[profileName] = &config.Profile{
		Default: def,
		Email:   email,
		Token:   token,
	}

	// If the user wants the newly created profile to be their new default, then
	// we'll call Set for its side effect of resetting all other profiles to have
	// their Default field set to false.
	if def {
		if p, ok := profile.Set(profileName, c.Globals.File.Profiles); ok {
			c.Globals.File.Profiles = p
		}
	}
}

func (c *CreateCommand) persistCfg() error {
	// TODO: The following directory checks should be encapsulated by the
	// File.Write() method as this chunk of code is duplicated in various places.
	// Consider consolidating with pkg/filesystem/directory.go
	// This function is itself duplicated in pkg/commands/profile/update.go
	dir := filepath.Dir(c.Globals.Path)
	fi, err := os.Stat(dir)
	switch {
	case err == nil && !fi.IsDir():
		return fmt.Errorf("config file path %s isn't a directory", dir)
	case err != nil && os.IsNotExist(err):
		if err := os.MkdirAll(dir, config.DirectoryPermissions); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Directory":   dir,
				"Permissions": config.DirectoryPermissions,
			})
			return fmt.Errorf("error creating config file directory: %w", err)
		}
	}

	if err := c.Globals.File.Write(c.Globals.Path); err != nil {
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
	text.Break(out)
	cont, err := text.AskYesNo(out, "Set this profile to be your default? [y/N] ", in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return false, err
	}
	text.Break(out)
	return cont, nil
}
