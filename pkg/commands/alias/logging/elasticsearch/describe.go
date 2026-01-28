package elasticsearch

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/elasticsearch"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the newcmd package.
type DescribeCommand struct {
	*newcmd.DescribeCommand
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{newcmd.NewDescribeCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service logging elasticsearch describe' command instead.")
	return c.DescribeCommand.Exec(in, out)
}
