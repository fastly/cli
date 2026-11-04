package resourcelink

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/resourcelink"

	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the newcmd package.
type DescribeCommand struct {
	newcmd.DescribeCommand
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service resource-link describe' command instead.")
	}
	return c.DescribeCommand.Exec(in, out)
}
