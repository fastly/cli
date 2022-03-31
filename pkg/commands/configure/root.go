package configure

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// APIClientFactory allows the configure command to regenerate the global Fastly
// API client when a new token is provided, in order to validate that token.
// It's a redeclaration of the app.APIClientFactory to avoid an import loop.
type APIClientFactory func(token, endpoint string) (api.Interface, error)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	clientFactory APIClientFactory
	delete        bool
	display       bool
	location      bool
	profiles      bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, cf APIClientFactory, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("configure", "Configure the Fastly CLI")
	c.CmdClause.Flag("delete", "Delete the specified --profile").Short('x').BoolVar(&c.delete)
	c.CmdClause.Flag("display", "Print the CLI configuration file").Short('d').BoolVar(&c.display)
	c.CmdClause.Flag("location", "Print the location of the CLI configuration file").Short('l').BoolVar(&c.location)
	c.CmdClause.Flag("profiles", "Print the available profiles").Short('p').BoolVar(&c.profiles)
	c.clientFactory = cf
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.profiles {
		return c.displayProfiles(out)
	}
	if c.Globals.Flag.Profile == "" && c.delete {
		msg := "--profile flag not provided"
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: "Provide both the --profile and --delete flags.",
		}
	}
	if c.Globals.Flag.Profile != "" && c.delete {
		return c.deleteProfile(out)
	}
	if c.location || c.display {
		return c.cfg()
	}

	if c.Globals.Flag.Profile != "" {
		if err = c.switchProfile(in, out); err != nil {
			return err
		}
		displayCfgPath(c.Globals.Path, out)
		text.Success(out, "Profile switched successfully to '%s'", c.Globals.Flag.Profile)
		return nil
	}

	name, err := c.promptForName(in, out)
	if err != nil {
		return err
	}
	if n, _ := profile.Get(name, c.Globals.File.Profiles); n != "" {
		return fmt.Errorf("profile '%s' already exists", n)
	}
	if err := c.tokenFlow(name, in, out); err != nil {
		return err
	}
	if err := c.persistCfg(); err != nil {
		return err
	}

	displayCfgPath(c.Globals.Path, out)
	text.Success(out, "Profile '%s' created", name)
	return nil
}

func (c *RootCommand) displayProfiles(out io.Writer) error {
	if c.Globals.File.Profiles == nil {
		msg := "no profiles available"
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	for k, v := range c.Globals.File.Profiles {
		text.Break(out)
		text.Output(out, text.Bold(k))
		text.Break(out)
		text.Output(out, "%s: %t", text.Bold("Default"), v.Default)
		text.Output(out, "%s: %s", text.Bold("Email"), v.Email)
		text.Output(out, "%s: %s", text.Bold("Token"), v.Token)
	}
	return nil
}

func (c *RootCommand) deleteProfile(out io.Writer) error {
	if ok := profile.Delete(c.Globals.Flag.Profile, c.Globals.File.Profiles); ok {
		if err := c.Globals.File.Write(c.Globals.Path); err != nil {
			return err
		}
		text.Success(out, "The profile '%s' was deleted.", c.Globals.Flag.Profile)

		if len(c.Globals.File.Profiles) > 0 {
			text.Warning(out, "At least one account profile should be set as the 'default'. Run `fastly configure --profile <NAME>`.")
		}
		return nil
	}
	return fmt.Errorf("the specified profile does not exist")
}

// cfg handles displaying the config data and file location.
func (c *RootCommand) cfg() error {
	if c.location && c.display {
		fmt.Printf("\n%s\n\n%s\n", text.Bold("LOCATION"), config.FilePath)

		data, err := os.ReadFile(config.FilePath)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		fmt.Printf("\n%s\n\n%s\n", text.Bold("CONFIG"), string(data))

		return nil
	}

	if c.location {
		fmt.Println(config.FilePath)
		return nil
	}

	if c.display {
		data, err := os.ReadFile(config.FilePath)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

// switchProfile updates config in memory and on disk to use specified profile.
func (c *RootCommand) switchProfile(in io.Reader, out io.Writer) error {
	var ok bool

	if c.Globals.File.Profiles, ok = profile.Set(c.Globals.Flag.Profile, c.Globals.File.Profiles); !ok {
		msg := fmt.Sprintf("A new profile '%s' will now be generated.", c.Globals.Flag.Profile)
		text.Warning(out, fmt.Sprintf("%s %s", profile.DoesNotExist, msg))

		if err := c.tokenFlow(c.Globals.Flag.Profile, in, out); err != nil {
			return err
		}
	}

	return c.persistCfg()
}

// tokenFlow initialises the token flow.
func (c *RootCommand) tokenFlow(profile string, in io.Reader, out io.Writer) error {
	var err error

	progress := text.NewProgress(out, c.Globals.Verbose())
	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
	}()

	// If user provides --token flag, then don't prompt them for input.
	token, source := c.Globals.Token()
	if source == config.SourceFile || source == config.SourceUndefined {
		token, err = promptForToken(in, out, c.Globals.ErrLog)
		if err != nil {
			return err
		}
		text.Break(out)
		text.Break(out)
	}

	endpoint, _ := c.Globals.Endpoint()

	user, err := c.validateToken(token, endpoint, progress)
	if err != nil {
		return err
	}

	c.updateInMemCfg(profile, user.Login, token, endpoint, progress)

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
// token in the configure command.
var ErrEmptyToken = errors.New("token cannot be empty")

// validateToken ensures the token can be used to acquire user data.
func (c *RootCommand) validateToken(token, endpoint string, progress text.Progress) (*fastly.User, error) {
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

func (c *RootCommand) updateInMemCfg(profileName, email, token, endpoint string, progress text.Progress) {
	progress.Step("Persisting configuration...")

	c.Globals.File.Fastly.APIEndpoint = endpoint

	if c.Globals.File.Profiles == nil {
		c.Globals.File.Profiles = make(config.Profiles)
	}
	c.Globals.File.Profiles[profileName] = &config.Profile{
		Default: true,
		Email:   email,
		Token:   token,
	}

	// We call Set for its side effect of resetting all other profiles to have
	// their Default field set to false.
	if p, ok := profile.Set(profileName, c.Globals.File.Profiles); ok {
		c.Globals.File.Profiles = p
	}
}

func (c *RootCommand) persistCfg() error {
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

func (c *RootCommand) promptForName(in io.Reader, out io.Writer) (string, error) {
	text.Break(out)

	var msg string
	if profile, _ := profile.Default(c.Globals.File.Profiles); profile == "" {
		msg = "No existing profiles exist. A new profile will be created."
	} else {
		msg = fmt.Sprintf("A default profile already exists (%s).\nA new default profile will now be configured.", profile)
	}

	text.Output(out, msg)
	text.Break(out)
	name, err := text.Input(out, "Profile name: ", in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return "", err
	}
	if name == "" {
		return "", fsterr.ErrNoProfileInput
	}
	text.Break(out)
	return name, nil
}
