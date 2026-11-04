package purge

import (
	"io"

	servicepurge "github.com/fastly/cli/pkg/commands/service/purge"

	"github.com/fastly/cli/pkg/text"
)

// PurgeCommand wraps the PurgeCommand from the servicepurge package.
type Command struct {
	servicepurge.PurgeCommand
}

// Exec implements the command interface.
func (c *Command) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service purge' command instead.")
	return c.PurgeCommand.Exec(in, out)
}
