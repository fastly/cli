package custom

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/vcl/custom"

	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the newcmd package.
type DescribeCommand struct {
	newcmd.DescribeCommand
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service vcl custom describe' command instead.")
	return c.DescribeCommand.Exec(in, out)
}
