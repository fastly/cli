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
	c.CmdClause.Arg("name", "Name of the token to remove").Required().StringVar(&c.name)
	return &c
}

func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.Globals.Config.DeleteAuthToken(c.name) {
		return fmt.Errorf("token %q not found", c.name)
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Token %q removed", c.name)
	return nil
}
