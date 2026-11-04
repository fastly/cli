package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/text"
)

// CloneCommand wraps the CloneCommand from the newcmd package.
type CloneCommand struct {
	newcmd.CloneCommand
}

// Exec implements the command interface.
func (c *CloneCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service version clone' command instead.")
	}
	return c.CloneCommand.Exec(in, out)
}
