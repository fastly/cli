package imageoptimizerdefaults

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/imageoptimizerdefaults"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand wraps the GetCommand from the newcmd package.
type GetCommand struct {
	*newcmd.GetCommand
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{newcmd.NewGetCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *GetCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service imageoptimizerdefaults get' command instead.")
	return c.GetCommand.Exec(in, out)
}
