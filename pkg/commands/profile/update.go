package profile

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand represents a Kingpin command.
type UpdateCommand struct {
	argparser.Base

	profile string
	sso     bool
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = g
	c.CmdClause = parent.Command("update", "Update user profile (deprecated: use 'fastly auth login' or 'fastly auth add' instead)")
	c.CmdClause.Arg("profile", "Profile to update (defaults to the currently active profile)").Short('p').StringVar(&c.profile)
	c.CmdClause.Flag("sso", "Update profile to use an SSO-based token").BoolVar(&c.sso)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth login' or 'fastly auth add' instead.\n\n")

	profileName, at, err := c.identifyProfile()
	if err != nil {
		return fmt.Errorf("failed to identify the profile to update: %w", err)
	}
	if c.Globals.Verbose() {
		text.Break(out)
	}
	text.Info(out, "Profile being updated: '%s'.\n\n", profileName)

	if err := c.updateToken(profileName, at, in, out); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	makeDefault := true
	if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Break(out)
		makeDefault, err = text.AskYesNo(out, text.BoldYellow("Make profile the default? [y/N] "), in)
		text.Break(out)
		if err != nil {
			return err
		}
	}

	if makeDefault {
		if err := c.Globals.Config.SetDefaultAuthToken(profileName); err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}
		if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error saving config file: %w", err)
		}
	}

	text.Success(out, "\nProfile '%s' updated", profileName)
	return nil
}

func (c *UpdateCommand) identifyProfile() (string, *config.AuthToken, error) {
	if c.profile == "" && c.Globals.Flags.Profile == "" {
		name, at := c.Globals.Config.GetDefaultAuthToken()
		if at == nil {
			return "", nil, fsterr.RemediationError{
				Inner:       fmt.Errorf("no active profile"),
				Remediation: "At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>` and ensure the profile is set to be the default.",
			}
		}
		return name, at, nil
	}

	profileName := c.profile
	if c.Globals.Flags.Profile != "" {
		profileName = c.Globals.Flags.Profile
	}
	at := c.Globals.Config.GetAuthToken(profileName)
	if at == nil {
		msg := fmt.Sprintf("the profile '%s' does not exist", profileName)
		return "", nil, fsterr.RemediationError{
			Inner:       errors.New(msg),
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	return profileName, at, nil
}

func (c *UpdateCommand) updateToken(profileName string, at *config.AuthToken, in io.Reader, out io.Writer) error {
	if c.sso || at.Type == config.AuthTokenTypeSSO {
		if err := authcmd.RunSSOWithTokenName(in, out, c.Globals, false, false, profileName); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		text.Break(out)
		return nil
	}

	token, err := text.InputSecure(out, text.BoldYellow("Profile token: (leave blank to skip): "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if token == "" {
		token = at.Token
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

	authcmd.BuildAndStoreStaticToken(c.Globals, token, profileName, md, false)

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}
	return nil
}
