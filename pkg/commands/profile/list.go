package profile

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand represents a Kingpin command.
type ListCommand struct {
	cmd.Base
	json bool
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.Globals = g
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
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(c.Globals.Config.Profiles)
			if err != nil {
				return err
			}
			_, err = out.Write(data)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return fmt.Errorf("error: unable to write data to stdout: %w", err)
			}
			return nil
		}
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
	if name == "" {
		text.Warning(out, profile.NoDefaults)
	} else {
		text.Info(out, "Default profile highlighted in red.")
		display(name, p, out, text.BoldRed)
	}

	for k, v := range c.Globals.Config.Profiles {
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
