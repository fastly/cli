package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the servicehealthcheck package.
type ListCommand struct {
	servicehealthcheck.ListCommand
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service healthcheck list' command instead.")
	}
	return c.ListCommand.Exec(in, out)
}
