package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the servicehealthcheck package.
type CreateCommand struct {
	servicehealthcheck.CreateCommand
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service healthcheck create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
