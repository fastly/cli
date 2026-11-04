package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the servicehealthcheck package.
type DescribeCommand struct {
	servicehealthcheck.DescribeCommand
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service healthcheck describe' command instead.")
	}
	return c.DescribeCommand.Exec(in, out)
}
