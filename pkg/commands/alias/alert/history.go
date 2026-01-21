package alert

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/alert"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListHistoryCommand wraps the ListHistoryCommand from the newcmd package.
type ListHistoryCommand struct {
	*newcmd.ListHistoryCommand
}

// NewListCommand returns a usable command registered under the parent.
func NewListHistoryCommand(parent argparser.Registerer, g *global.Data) *ListHistoryCommand {
	c := ListHistoryCommand{newcmd.NewListHistoryCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *ListHistoryCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service alert list history' command instead.")
	return c.ListHistoryCommand.Exec(in, out)
}
