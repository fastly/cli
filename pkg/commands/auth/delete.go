package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand removes a stored token.
type DeleteCommand struct {
	argparser.Base
	// Required.
	NameFlag string `arg:"" name:"name" required:"" help:"Name of the token to remove"`
}

func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Config.GetAuthToken(c.NameFlag) == nil {
		return fmt.Errorf("token %q not found", c.NameFlag)
	}

	wasDefault := c.Globals.Config.Auth.Default == c.NameFlag

	if wasDefault && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Warning(out, "%q is your current default token. Deleting it will affect commands that don't use --token or FASTLY_API_TOKEN.", c.NameFlag)
		cont, err := text.AskYesNo(out, "Are you sure? [y/N]: ", in)
		if err != nil {
			return err
		}
		if !cont {
			return nil
		}
	}

	c.Globals.Config.DeleteAuthToken(c.NameFlag)

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Token %q removed", c.NameFlag)
	if wasDefault {
		if c.Globals.Config.Auth.Default != "" {
			text.Info(out, "Default token reassigned to %q", c.Globals.Config.Auth.Default)
		} else {
			text.Warning(out, "No default token configured; use 'fastly auth use <name>' to set one")
		}
	}
	return nil
}
