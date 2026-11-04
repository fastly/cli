package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// DeactivateCommand wraps the DeactivateCommand from the newcmd package.
type DeactivateCommand struct {
	newcmd.DeactivateCommand
}

// Exec implements the command interface.
func (c *DeactivateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service version deactivate' command instead.")
	return c.DeactivateCommand.Exec(in, out)
}
