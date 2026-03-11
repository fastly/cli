package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UseCommand switches the default token.
type UseCommand struct {
	argparser.Base
	name string
}

func NewUseCommand(parent argparser.Registerer, g *global.Data) *UseCommand {
	var c UseCommand
	c.Globals = g
	c.CmdClause = parent.Command("use", "Set the default stored token for CLI commands")
	// Required.
	c.CmdClause.Arg("name", "Name of the token to use as default").Required().StringVar(&c.name)
	return &c
}

func (c *UseCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := c.Globals.Config.SetDefaultAuthToken(c.name); err != nil {
		return err
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Default token switched to %q", c.name)
	return nil
}
