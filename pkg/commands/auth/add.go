package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AddCommand adds a named token entry.
type AddCommand struct {
	argparser.Base
	// Optional.
	NameFlag string `arg:"" name:"name" optional:"" help:"Name for this token (pass to --token to use it later); if omitted, uses the API token's name"`
	// Required.
	Token string `name:"api-token" required:"" help:"Fastly API token to store"`
}

func (c *AddCommand) Exec(_ io.Reader, out io.Writer) error {
	// Short-circuit: if an explicit name was provided and already exists,
	// fail before making any network calls.
	if c.NameFlag != "" && c.Globals.Config.GetAuthToken(c.NameFlag) != nil {
		return fmt.Errorf("token %q already exists; use 'fastly auth delete %s' first", c.NameFlag, c.NameFlag)
	}

	md, err := FetchTokenMetadata(c.Globals, c.Token)
	if err != nil {
		return err
	}

	name := c.NameFlag
	if name == "" {
		if md.APITokenName == "" {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("could not determine a name for this token"),
				Remediation: "Provide a name as the first argument, e.g.: fastly auth add my-token --api-token <token>",
			}
		}
		name = md.APITokenName
		// Check collision for the derived name too.
		if c.Globals.Config.GetAuthToken(name) != nil {
			return fmt.Errorf("token %q already exists; use 'fastly auth delete %s' first", name, name)
		}
	}

	entry := &config.AuthToken{
		Type:              config.AuthTokenTypeStatic,
		Token:             c.Token,
		Email:             md.Email,
		AccountID:         md.AccountID,
		APITokenName:      md.APITokenName,
		APITokenScope:     md.APITokenScope,
		APITokenExpiresAt: md.APITokenExpiresAt,
		APITokenID:        md.APITokenID,
	}

	c.Globals.Config.SetAuthToken(name, entry)

	// When no default token is configured, automatically promote this token
	// so CLI commands work without an explicit --token flag.
	setDefault := c.Globals.Config.Auth.Default == ""
	if setDefault {
		c.Globals.Config.Auth.Default = name
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Token %q added", name)
	if setDefault {
		text.Info(out, "Token %q set as default (no previous default was configured)", name)
	}
	text.Info(out, "Token saved to %s", c.Globals.ConfigPath)
	return nil
}
