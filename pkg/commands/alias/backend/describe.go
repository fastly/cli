package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the servicebackend package.
type DescribeCommand struct {
	servicebackend.DescribeCommand
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service backend describe' command instead.")
	}
	return c.DescribeCommand.Exec(in, out)
}
