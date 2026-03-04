package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// LoginCommand stores a token as the default credential.
type LoginCommand struct {
	argparser.Base
	sso bool
}

func NewLoginCommand(parent argparser.Registerer, g *global.Data) *LoginCommand {
	var c LoginCommand
	c.Globals = g
	c.CmdClause = parent.Command("login", "Authenticate and store a default token (paste token or use --sso)")
	// Optional.
	c.CmdClause.Flag("sso", "Authenticate via browser-based SSO (requires --token <name> to specify the stored token name)").BoolVar(&c.sso)
	return &c
}

func (c *LoginCommand) Exec(in io.Reader, out io.Writer) error {
	if c.sso {
		return c.execSSO(in, out)
	}

	text.Output(out, "An API token can be generated at: https://manage.fastly.com/account/personal/tokens\n\n")

	token, err := text.InputSecure(out, "Paste your API token: ", in)
	if err != nil {
		return fmt.Errorf("error reading token input: %w", err)
	}
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	name, md, err := StoreStaticToken(c.Globals, token)
	if err != nil {
		return err
	}

	text.Success(out, "Authenticated as %s (token stored as %q)", md.Email, name)
	text.Info(out, "Token saved to %s", c.Globals.ConfigPath)
	return nil
}

func (c *LoginCommand) execSSO(in io.Reader, out io.Writer) error {
	if c.Globals.Flags.Token == "" {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("SSO login requires a token name via --token"),
			Remediation: "Provide a name for the stored token, e.g.: fastly auth login --sso --token work-sso",
		}
	}
	tokenName := c.Globals.Flags.Token

	if c.Globals.AuthServer == nil {
		return fmt.Errorf("SSO authentication requires network access to the Fastly OIDC provider, but the auth server could not be configured; use 'fastly auth login' (without --sso) to paste a static API token instead")
	}

	if err := RunSSOWithTokenName(in, out, c.Globals, false, false, tokenName); err != nil {
		return fmt.Errorf("SSO authentication failed: %w", err)
	}

	c.Globals.Config.Auth.Default = tokenName
	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Break(out)
	text.Success(out, "Authenticated via SSO.")
	return nil
}
