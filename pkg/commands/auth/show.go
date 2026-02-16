package auth

import (
	"fmt"
	"io"
	"time"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
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
	now := time.Now()
	status, expires, parseErr := GetExpirationStatus(entry, now)
	if parseErr != nil && c.Globals.ErrLog != nil {
		c.Globals.ErrLog.Add(parseErr)
	}

	if entry.APITokenExpiresAt != "" {
		line := "API token expires at: " + entry.APITokenExpiresAt
		if summary := apiTokenExpirySummary(entry, expires, now); summary != "" {
			line += " (" + summary + ")"
		}
		text.Output(out, "%s\n", line)
	}
	if entry.APITokenID != "" {
		text.Output(out, "API token ID: %s\n", entry.APITokenID)
	}

	// For SSO tokens, show the session (refresh) expiry as the user-actionable deadline.
	if entry.Type == config.AuthTokenTypeSSO && entry.RefreshExpiresAt != "" && !entry.NeedsReauth {
		line := "SSO session expires at: " + entry.RefreshExpiresAt
		if summary := ExpirationSummary(status, expires, now); summary != "" {
			line += " (" + summary + ")"
		}
		text.Output(out, "%s\n", line)
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
		text.Warning(out, "This token needs re-authentication. %s\n", ExpirationRemediation(entry.Type))
	} else if status == StatusExpired {
		text.Warning(out, "This token has expired. %s\n", ExpirationRemediation(entry.Type))
	}

	return nil
}

// apiTokenExpirySummary returns a relative-time string for the APITokenExpiresAt
// field specifically. For static tokens this uses the main expiry status; for SSO
// tokens the APITokenExpiresAt is secondary so we parse it independently.
func apiTokenExpirySummary(entry *config.AuthToken, mainExpires time.Time, now time.Time) string {
	if entry.Type == config.AuthTokenTypeStatic {
		// For static tokens, APITokenExpiresAt IS the effective expiry.
		if mainExpires.IsZero() {
			return ""
		}
		if now.After(mainExpires) {
			return "expired " + humanDuration(now.Sub(mainExpires)) + " ago"
		}
		return "in " + humanDuration(mainExpires.Sub(now))
	}

	// For SSO tokens, parse APITokenExpiresAt independently since the main
	// expiry status tracks RefreshExpiresAt.
	if entry.APITokenExpiresAt == "" {
		return ""
	}
	apiExpires, err := time.Parse(time.RFC3339, entry.APITokenExpiresAt)
	if err != nil {
		return ""
	}
	if now.After(apiExpires) {
		return "expired " + humanDuration(now.Sub(apiExpires)) + " ago"
	}
	return "in " + humanDuration(apiExpires.Sub(now))
}
