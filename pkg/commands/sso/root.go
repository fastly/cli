package sso

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ForceReAuth indicates we want to force a re-auth of the user's session.
// This variable is overridden by ../../app/run.go to force a re-auth.
var ForceReAuth = false

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	profile string
}

// CommandName is the string to be used to invoke this command.
const CommandName = "sso"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Single Sign-On authentication (deprecated: use 'fastly auth login --sso' instead)").Hidden()
	c.CmdClause.Arg("profile", "Profile to authenticate (i.e. create/update a token for)").Short('p').StringVar(&c.profile)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth login --sso' instead.\n\n")

	tokenName, isFallback := c.resolveTokenName()

	if !isFallback && c.Globals.Config.GetAuthToken(tokenName) == nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("token %q does not exist", tokenName),
			Remediation: "Run 'fastly auth login --sso' to create a new SSO token, or 'fastly auth add' to store an existing token.",
		}
	}

	if err := authcmd.RunSSOWithTokenName(in, out, c.Globals, ForceReAuth, false, tokenName); err != nil {
		return err
	}

	c.Globals.Config.Auth.Default = tokenName
	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}
	return nil
}

func (c *RootCommand) resolveTokenName() (string, bool) {
	if c.Globals.Flags.Profile != "" {
		return c.Globals.Flags.Profile, false
	}
	if c.Globals.Manifest != nil && c.Globals.Manifest.File.Profile != "" {
		return c.Globals.Manifest.File.Profile, false
	}
	if c.profile != "" {
		return c.profile, false
	}
	if name, _ := c.Globals.Config.GetDefaultAuthToken(); name != "" {
		return name, false
	}
	return "default", true
}
