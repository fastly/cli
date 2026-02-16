package profile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/argparser"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand represents a Kingpin command.
type CreateCommand struct {
	argparser.Base

	profile string
	sso     bool
}

// NewCreateCommand returns a new command registered in the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = g
	c.CmdClause = parent.Command("create", "Create user profile (deprecated: use 'fastly auth login' or 'fastly auth add' instead)")
	c.CmdClause.Arg("profile", "Profile to create (default 'user')").Default("user").Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("sso", "Create an SSO-based token").BoolVar(&c.sso)
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) (err error) {
	text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth login' or 'fastly auth add' instead.\n\n")

	if c.Globals.Verbose() {
		text.Break(out)
	}
	text.Output(out, "Creating profile '%s'", c.profile)

	if c.Globals.Config.GetAuthToken(c.profile) != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("profile '%s' already exists", c.profile),
			Remediation: "Re-run the command and pass a different value for the 'profile' argument.",
		}
	}

	makeDefault := true
	if name, _ := c.Globals.Config.GetDefaultAuthToken(); name != "" && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		makeDefault, err = c.promptForDefault(in, out)
		if err != nil {
			return err
		}
	}
	text.Break(out)

	if c.sso {
		if err := authcmd.RunSSOWithTokenName(in, out, c.Globals, false, false, c.profile); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		if makeDefault {
			c.Globals.Config.Auth.Default = c.profile
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

func (c *CreateCommand) staticTokenFlow(makeDefault bool, in io.Reader, out io.Writer) error {
	token, err := promptForToken(in, out, c.Globals.ErrLog)
	if err != nil {
		return err
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

	var md *authcmd.TokenMetadata
	err = spinner.Process("Validating token", func(_ *text.SpinnerWrapper) error {
		md, err = authcmd.FetchTokenMetadata(c.Globals, token)
		return err
	})
	if err != nil {
		return err
	}

	return spinner.Process("Persisting configuration", func(_ *text.SpinnerWrapper) error {
		authcmd.BuildAndStoreStaticToken(c.Globals, token, c.profile, md, makeDefault)
		return nil
	})
}

func promptForToken(in io.Reader, out io.Writer, errLog fsterr.LogInterface) (string, error) {
	text.Output(out, "An API token is used to authenticate requests to the Fastly API. To create a token, visit https://manage.fastly.com/account/personal/tokens\n\n")
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

func (c *CreateCommand) persistCfg() error {
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
	cont, err := text.AskYesNo(out, "\nSet this profile to be your default? [y/N] ", in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return false, err
	}
	return cont, nil
}
