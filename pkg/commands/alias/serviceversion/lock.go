package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// LockCommand wraps the LockCommand from the newcmd package.
type LockCommand struct {
	*newcmd.LockCommand
}

// NewLockCommand returns a usable command registered under the parent.
func NewLockCommand(parent argparser.Registerer, g *global.Data) *LockCommand {
	c := LockCommand{newcmd.NewLockCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *LockCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service version lock' command instead.")
	return c.LockCommand.Exec(in, out)
}
