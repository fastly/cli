package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UnstageCommand wraps the UnstageCommand from the newcmd package.
type UnstageCommand struct {
	*newcmd.UnstageCommand
}

// NewUnstageCommand returns a usable command registered under the parent.
func NewUnstageCommand(parent argparser.Registerer, g *global.Data) *UnstageCommand {
	c := UnstageCommand{newcmd.NewUnstageCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *UnstageCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service version unstage' command instead.")
	return c.UnstageCommand.Exec(in, out)
}
