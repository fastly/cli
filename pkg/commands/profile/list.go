package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand represents a Kingpin command.
type ListCommand struct {
	cmd.Base
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List user profiles")
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.File.Profiles == nil {
		msg := "no profiles available"
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	if len(c.Globals.File.Profiles) == 0 {
		text.Break(out)
		text.Description(out, "No profiles defined. To create a profile, run", "fastly profile create <name>")
		return nil
	}
	if name, _ := profile.Default(c.Globals.File.Profiles); name == "" {
		text.Warning(out, profile.NoDefaults)
	} else {
		text.Info(out, "Default profile highlighted in red.")
	}

	for k, v := range c.Globals.File.Profiles {
		style := text.Bold
		if v.Default {
			style = text.BoldRed
		}

		text.Break(out)
		text.Output(out, style(k))
		text.Break(out)
		text.Output(out, "%s: %t", style("Default"), v.Default)
		text.Output(out, "%s: %s", style("Email"), v.Email)
		text.Output(out, "%s: %s", style("Token"), v.Token)
	}
	return nil
}
