package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the servicebackend package.
type ListCommand struct {
	servicebackend.ListCommand
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service backend list' command instead.")
	}
	return c.ListCommand.Exec(in, out)
}
