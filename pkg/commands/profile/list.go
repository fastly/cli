package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
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
	c.CmdClause = parent.Command("list", "List user profiles")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if ok, err := c.WriteJSON(out, c.Globals.Config.Profiles); ok {
		return err
	}

	if c.Globals.Config.Profiles == nil {
		msg := "no profiles available"
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	if len(c.Globals.Config.Profiles) == 0 {
		text.Break(out)
		text.Description(out, "No profiles defined. To create a profile, run", "fastly profile create <name>")
		return nil
	}

	name, p := profile.Default(c.Globals.Config.Profiles)
	if p == nil {
		text.Warning(out, profile.NoDefaults)
	} else {
		text.Info(out, "Default profile highlighted in red.\n\n")
		display(name, p, out, text.BoldRed)
	}

	for k, v := range c.Globals.Config.Profiles {
		if !v.Default {
			text.Break(out)
			display(k, v, out, text.Bold)
		}
	}
	return nil
}

func display(k string, v *config.Profile, out io.Writer, style func(a ...any) string) {
	text.Output(out, style(k))
	text.Break(out)
	text.Output(out, "%s: %t", style("Default"), v.Default)
	text.Output(out, "%s: %s", style("Email"), v.Email)
	text.Output(out, "%s: %s", style("Token"), v.Token)
	text.Output(out, "%s: %t", style("SSO"), !auth.IsLongLivedToken(v))
}
