package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/sso"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
)

// RootCommand is the setup command.
type RootCommand struct {
	argparser.Base
	ssoCmd *sso.RootCommand

	profileName string
	setDefault  argparser.OptionalBool
}

// NewRootCommand creates a new setup command.
func NewRootCommand(parent argparser.Registerer, g *global.Data, ssoCmd *sso.RootCommand) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.ssoCmd = ssoCmd
	c.CmdClause = parent.Command("setup", "Interactive setup wizard for configuring the Fastly CLI")
	c.CmdClause.Flag("name", "Profile name to create").Default(profile.DefaultName).StringVar(&c.profileName)
	c.CmdClause.Flag("set-default", "Set as default profile").Action(c.setDefault.Set).BoolVar(&c.setDefault.Value)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	token := c.Globals.Flags.Token
	nonInteractive := c.Globals.Flags.NonInteractive
	autoYes := c.Globals.Flags.AutoYes

	if nonInteractive && token == "" {
		return fsterr.RemediationError{
			Inner:       errors.New("--token is required when using --non-interactive"),
			Remediation: "Provide an API token: fastly setup --non-interactive --token $FASTLY_API_TOKEN",
		}
	}

	if profile.Exist(c.profileName, c.Globals.Config.Profiles) {
		if nonInteractive {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("profile '%s' already exists", c.profileName),
				Remediation: "Use 'fastly profile update' to modify, or specify a different --name.",
			}
		}

		if autoYes {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("profile '%s' already exists", c.profileName),
				Remediation: "Specify a different profile name: fastly setup -y --name <new-name>",
			}
		}

		text.Warning(out, "A profile named '%s' already exists.", c.profileName)
		cont, err := text.AskYesNo(out, "Would you like to create a profile with a different name? [y/N] ", in)
		if err != nil {
			return err
		}
		if !cont {
			text.Info(out, "Setup cancelled.")
			return nil
		}
	}

	if nonInteractive {
		return c.runNonInteractive(out, token, nonInteractive)
	}
	return c.runInteractive(in, out, autoYes)
}

// verifyResponse models the /verify endpoint response.
type verifyResponse struct {
	Customer struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"customer"`
	User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	} `json:"user"`
	Token struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Scope     string `json:"scope"`
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at"`
	} `json:"token"`
}

func (c *RootCommand) validateToken(token string) (*verifyResponse, error) {
	endpoint, _ := c.Globals.APIEndpoint()
	data, err := undocumented.Call(undocumented.CallOptions{
		APIEndpoint: endpoint,
		HTTPClient:  c.Globals.HTTPClient,
		HTTPHeaders: []undocumented.HTTPHeader{
			{Key: "Accept", Value: "application/json"},
			{Key: "User-Agent", Value: useragent.Name},
		},
		Method: http.MethodGet,
		Path:   "/verify",
		Token:  token,
	})
	if err != nil {
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	var resp verifyResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("error decoding API response: %w", err)
	}
	return &resp, nil
}

func (c *RootCommand) runNonInteractive(out io.Writer, token string, _ bool) error {
	resp, err := c.validateToken(token)
	if err != nil {
		return err
	}

	makeDefault := c.setDefault.Value
	if !c.setDefault.WasSet {
		_, defaultProfile := profile.Default(c.Globals.Config.Profiles)
		makeDefault = (defaultProfile == nil)
	}

	if err := c.createProfile(c.profileName, token, resp, makeDefault); err != nil {
		return err
	}

	if makeDefault {
		text.Success(out, "Profile '%s' created and set as default.", c.profileName)
	} else {
		text.Success(out, "Profile '%s' created.", c.profileName)
	}
	return nil
}

func (c *RootCommand) runInteractive(in io.Reader, out io.Writer, autoYes bool) error {
	text.Output(out, "Welcome to Fastly CLI Setup!")
	text.Break(out)
	text.Output(out, "This wizard will help you configure authentication and create a profile.")
	text.Break(out)

	useSSO := false
	if !autoYes {
		text.Output(out, "How would you like to authenticate?")
		text.Break(out)
		text.Output(out, "  1. API Token (recommended for automation)")
		text.Output(out, "  2. SSO/Browser Login (recommended for interactive use)")
		text.Break(out)
		choice, err := text.Input(out, "Choice [1]: ", in)
		if err != nil {
			return err
		}
		useSSO = (choice == "2")
	}

	profileName := c.profileName
	needsNewName := profile.Exist(c.profileName, c.Globals.Config.Profiles)

	if !autoYes {
		prompt := fmt.Sprintf("Profile name [%s]: ", c.profileName)
		if needsNewName {
			prompt = "Enter a new profile name: "
		}

		for {
			input, err := text.Input(out, prompt, in)
			if err != nil {
				return err
			}
			input = strings.TrimSpace(input)

			if input == "" && !needsNewName {
				break
			}

			if input == "" {
				text.Warning(out, "Profile name cannot be empty.")
				continue
			}

			if profile.Exist(input, c.Globals.Config.Profiles) {
				text.Warning(out, "Profile '%s' already exists. Please choose a different name.", input)
				needsNewName = true
				prompt = "Enter a new profile name: "
				continue
			}

			profileName = input
			break
		}
	} else if needsNewName {
		return fmt.Errorf("profile '%s' already exists", c.profileName)
	}

	_, defaultProfile := profile.Default(c.Globals.Config.Profiles)
	hasExistingDefault := (defaultProfile != nil)

	makeDefault := true
	if c.setDefault.WasSet {
		makeDefault = c.setDefault.Value
	} else if hasExistingDefault {
		if autoYes {
			makeDefault = false
		} else {
			var err error
			makeDefault, err = text.AskYesNo(out, "Set this profile as your default? [y/N] ", in)
			if err != nil {
				return err
			}
		}
	}

	if useSSO {
		return c.runSSOFlow(in, out, profileName, makeDefault)
	}
	return c.runAPITokenFlow(in, out, profileName, makeDefault)
}

func (c *RootCommand) runSSOFlow(in io.Reader, out io.Writer, profileName string, makeDefault bool) error {
	c.ssoCmd.InvokedFromProfileCreate = true
	c.ssoCmd.ProfileCreateName = profileName
	c.ssoCmd.ProfileDefault = makeDefault

	if err := c.ssoCmd.Exec(in, out); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	c.displaySummary(out, profileName, makeDefault)
	return nil
}

func (c *RootCommand) runAPITokenFlow(in io.Reader, out io.Writer, profileName string, makeDefault bool) error {
	text.Break(out)
	text.Output(out, "You can create an API token at: https://manage.fastly.com/account/personal/tokens")
	text.Break(out)
	token, err := text.InputSecure(out, "Fastly API token: ", in, profile.ValidateTokenNotEmpty)
	if err != nil {
		return err
	}
	if token == "" {
		return fsterr.RemediationError{
			Inner:       errors.New("API token cannot be empty"),
			Remediation: "Enter a valid API token, or create one at https://manage.fastly.com/account/personal/tokens",
		}
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	var resp *verifyResponse
	err = spinner.Process("Validating token", func(_ *text.SpinnerWrapper) error {
		resp, err = c.validateToken(token)
		return err
	})
	if err != nil {
		return err
	}

	text.Break(out)
	text.Success(out, "Token validated successfully!")
	fmt.Fprintf(out, "\n  Email: %s\n", resp.User.Login)
	fmt.Fprintf(out, "  Customer: %s (%s)\n", resp.Customer.Name, resp.Customer.ID)

	if err := c.createProfile(profileName, token, resp, makeDefault); err != nil {
		return err
	}

	c.displaySummary(out, profileName, makeDefault)
	return nil
}

func (c *RootCommand) createProfile(name, token string, resp *verifyResponse, makeDefault bool) error {
	c.Globals.Config.Profiles = profile.Create(
		name,
		c.Globals.Config.Profiles,
		resp.User.Login,
		token,
		makeDefault,
	)
	return c.Globals.Config.Write(c.Globals.ConfigPath)
}

func (c *RootCommand) displaySummary(out io.Writer, profileName string, isDefault bool) {
	text.Break(out)
	if isDefault {
		text.Success(out, "Setup complete! Profile '%s' created and set as default.", profileName)
	} else {
		text.Success(out, "Setup complete! Profile '%s' created.", profileName)
	}
	text.Description(out, "Your configuration has been saved to", c.Globals.ConfigPath)
	text.Break(out)
	text.Output(out, "Next steps:")
	text.Output(out, "  - Run 'fastly whoami' to verify your identity")
	text.Output(out, "  - Run 'fastly service list' to see your services")
	text.Output(out, "  - Run 'fastly compute init' to start a new Compute project")
}
