package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// UnstageCommand wraps the UnstageCommand from the newcmd package.
type UnstageCommand struct {
	newcmd.UnstageCommand
}

// Exec implements the command interface.
func (c *UnstageCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service version unstage' command instead.")
	return c.UnstageCommand.Exec(in, out)
}
