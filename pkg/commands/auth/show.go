package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/text"
)

// ShowCommand shows token details.
type ShowCommand struct {
	argparser.Base
	name   string
	reveal bool
}

func NewShowCommand(parent argparser.Registerer, g *global.Data) *ShowCommand {
	var c ShowCommand
	c.Globals = g
	c.CmdClause = parent.Command("show", "Show details for a stored token")
	c.CmdClause.Arg("name", "Name of the token to show (defaults to the current token)").StringVar(&c.name)
	c.CmdClause.Flag("reveal", "Show the full token value (use with care)").BoolVar(&c.reveal)
	return &c
}

func (c *ShowCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.name == "" {
		_, src := c.Globals.Token()
		switch src {
		case lookup.SourceUndefined:
			return fmt.Errorf("no token configured; run `fastly auth login` or pass a token name")
		case lookup.SourceFlag, lookup.SourceEnvironment:
			return fmt.Errorf("current token is not stored (provided via --token or %s); use `fastly auth add` or `fastly auth show <name>`", env.APIToken)
		case lookup.SourceFile, lookup.SourceDefault, lookup.SourceAuth:
			c.name = c.Globals.AuthTokenName()
			if c.name == "" {
				c.name = c.Globals.Config.Auth.Default
			}
		}
	}

	entry := c.Globals.Config.GetAuthToken(c.name)
	if entry == nil {
		return fmt.Errorf("token %q not found", c.name)
	}

	isDefault := c.name == c.Globals.Config.Auth.Default
	defaultStr := ""
	if isDefault {
		defaultStr = " (default)"
	}

	text.Output(out, "Name: %s%s\n", c.name, defaultStr)
	text.Output(out, "Type: %s\n", entry.Type)

	if entry.Email != "" {
		text.Output(out, "Email: %s\n", entry.Email)
	}
	if entry.AccountID != "" {
		text.Output(out, "Account ID: %s\n", entry.AccountID)
	}
	if entry.Label != "" {
		text.Output(out, "Label: %s\n", entry.Label)
	}
	if entry.APITokenName != "" {
		text.Output(out, "API token name: %s\n", entry.APITokenName)
	}
	if entry.APITokenScope != "" {
		text.Output(out, "API token scope: %s\n", entry.APITokenScope)
	}
	if entry.APITokenExpiresAt != "" {
		text.Output(out, "API token expires at: %s\n", entry.APITokenExpiresAt)
	}
	if entry.APITokenID != "" {
		text.Output(out, "API token ID: %s\n", entry.APITokenID)
	}

	if c.reveal {
		text.Output(out, "Token: %s\n", entry.Token)
	} else {
		if len(entry.Token) > 8 {
			text.Output(out, "Token: %s...%s\n", entry.Token[:4], entry.Token[len(entry.Token)-4:])
		} else {
			text.Output(out, "Token: ****\n")
		}
	}

	if entry.NeedsReauth {
		text.Warning(out, "This token needs re-authentication. Run `fastly auth login --sso` to refresh.\n")
	}

	return nil
}
