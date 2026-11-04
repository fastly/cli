package alerts

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/alert"

	"github.com/fastly/cli/pkg/text"
)

// ListHistoryCommand wraps the ListHistoryCommand from the newcmd package.
type ListHistoryCommand struct {
	newcmd.ListHistoryCommand
}

// Exec implements the command interface.
func (c *ListHistoryCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service alert list history' command instead.")
	}
	return c.ListHistoryCommand.Exec(in, out)
}
