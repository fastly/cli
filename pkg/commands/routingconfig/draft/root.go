package draft

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Diff   DiffCommand   `cmd:"" help:"Show the differences between a routing config's active and draft versions"`
	Update UpdateCommand `cmd:"" help:"Set the comment on a routing config's draft version"`
	Delete DeleteCommand `cmd:"" aliases:"discard" help:"Discard a routing config's draft version, reverting it back to the active version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "draft"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
