package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand removes a stored token.
type DeleteCommand struct {
	argparser.Base
	name string
}

func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = g
	c.CmdClause = parent.Command("delete", "Delete a stored token")
	// Required.
	c.CmdClause.Arg("name", "Name of the token to remove").Required().StringVar(&c.name)
	return &c
}

func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Config.GetAuthToken(c.name) == nil {
		return fmt.Errorf("token %q not found", c.name)
	}

	wasDefault := c.Globals.Config.Auth.Default == c.name

	if wasDefault && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		text.Warning(out, "%q is your current default token. Deleting it will affect commands that don't use --token or FASTLY_API_TOKEN.", c.name)
		cont, err := text.AskYesNo(out, "Are you sure? [y/N]: ", in)
		if err != nil {
			return err
		}
		if !cont {
			return nil
		}
	}

	c.Globals.Config.DeleteAuthToken(c.name)

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Token %q removed", c.name)
	if wasDefault {
		if c.Globals.Config.Auth.Default != "" {
			text.Info(out, "Default token reassigned to %q", c.Globals.Config.Auth.Default)
		} else {
			text.Warning(out, "No default token configured; use 'fastly auth use <name>' to set one")
		}
	}
	return nil
}
