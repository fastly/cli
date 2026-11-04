package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// StageCommand wraps the StageCommand from the newcmd package.
type StageCommand struct {
	newcmd.StageCommand
}

// Exec implements the command interface.
func (c *StageCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service version stage' command instead.")
	return c.StageCommand.Exec(in, out)
}
