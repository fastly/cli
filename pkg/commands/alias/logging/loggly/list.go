package loggly

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/loggly"

	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the newcmd package.
type ListCommand struct {
	newcmd.ListCommand
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service logging loggly list' command instead.")
	}
	return c.ListCommand.Exec(in, out)
}
