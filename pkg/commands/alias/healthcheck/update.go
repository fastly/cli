package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the servicehealthcheck package.
type UpdateCommand struct {
	servicehealthcheck.UpdateCommand
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service healthcheck update' command instead.")
	return c.UpdateCommand.Exec(in, out)
}
