package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the servicebackend package.
type DescribeCommand struct {
	*servicebackend.DescribeCommand
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{servicebackend.NewDescribeCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service backend describe' command instead.")
	return c.DescribeCommand.Exec(in, out)
}
