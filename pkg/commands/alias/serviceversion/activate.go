package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// ActivateCommand wraps the ActivateCommand from the newcmd package.
type ActivateCommand struct {
	newcmd.ActivateCommand
}

// Exec implements the command interface.
func (c *ActivateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service version activate' command instead.")
	return c.ActivateCommand.Exec(in, out)
}
