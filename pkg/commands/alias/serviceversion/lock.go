package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// LockCommand wraps the LockCommand from the newcmd package.
type LockCommand struct {
	newcmd.LockCommand
}

// Exec implements the command interface.
func (c *LockCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service version lock' command instead.")
	return c.LockCommand.Exec(in, out)
}
