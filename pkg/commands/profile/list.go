package profile

import (
	"encoding/json"
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
	json bool
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List user profiles")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(c.Globals.File.Profiles)
			if err != nil {
				return err
			}
			out.Write(data)
			return nil
		}
	}

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

	name, p := profile.Default(c.Globals.File.Profiles)
	if name == "" {
		text.Warning(out, profile.NoDefaults)
	} else {
		text.Info(out, "Default profile highlighted in red.")
		display(name, p, out, text.BoldRed)
	}

	for k, v := range c.Globals.File.Profiles {
		if !v.Default {
			display(k, v, out, text.Bold)
		}
	}
	return nil
}

func display(k string, v *config.Profile, out io.Writer, style func(a ...any) string) {
	text.Break(out)
	text.Output(out, style(k))
	text.Break(out)
	text.Output(out, "%s: %t", style("Default"), v.Default)
	text.Output(out, "%s: %s", style("Email"), v.Email)
	text.Output(out, "%s: %s", style("Token"), v.Token)
}
