package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CloneCommand wraps the CloneCommand from the newcmd package.
type CloneCommand struct {
	*newcmd.CloneCommand
}

// NewCloneCommand returns a usable command registered under the parent.
func NewCloneCommand(parent argparser.Registerer, g *global.Data) *CloneCommand {
	c := CloneCommand{newcmd.NewCloneCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *CloneCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated(out, "Use the 'service version clone' command instead.")
	}
	return c.CloneCommand.Exec(in, out)
}
