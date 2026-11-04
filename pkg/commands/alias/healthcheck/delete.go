package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the servicehealthcheck package.
type DeleteCommand struct {
	servicehealthcheck.DeleteCommand
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service healthcheck delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
