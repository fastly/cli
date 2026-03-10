package profile

import (
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand represents a Kingpin command.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.Globals = g
	c.CmdClause = parent.Command("list", "List user profiles (deprecated: use 'fastly auth list' instead)")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	text.Deprecated(out, "This command will be removed in a future release. Use 'fastly auth list' instead.\n\n")

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if ok, err := c.WriteJSON(out, c.Globals.Config.Auth.Tokens); ok {
		return err
	}

	if len(c.Globals.Config.Auth.Tokens) == 0 {
		msg := "no profiles available"
		return fsterr.RemediationError{
			Inner:       errors.New(msg),
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	defaultName := c.Globals.Config.Auth.Default

	if defaultName != "" {
		if at := c.Globals.Config.Auth.Tokens[defaultName]; at != nil {
			if c.Globals.Verbose() {
				text.Break(out)
			}
			text.Info(out, "Default profile highlighted in red.\n\n")
			display(defaultName, at, true, out, text.BoldRed)
		}
	}

	for name, at := range c.Globals.Config.Auth.Tokens {
		if name != defaultName {
			text.Break(out)
			display(name, at, false, out, text.Bold)
		}
	}
	return nil
}

func display(name string, at *config.AuthToken, isDefault bool, out io.Writer, style func(a ...any) string) {
	text.Output(out, style(name))
	text.Break(out)
	text.Output(out, "%s: %t", style("Default"), isDefault)
	text.Output(out, "%s: %s", style("Email"), at.Email)
	text.Output(out, "%s: %s", style("Token"), at.Token)
	isSSO := at.Type == config.AuthTokenTypeSSO
	text.Output(out, "%s: %t", style("SSO"), isSSO)
	if isSSO {
		text.Output(out, "%s: %s", style("Account ID"), at.AccountID)
		text.Output(out, "%s: %s", style("Label"), at.Label)
	}
}
