package tools

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/tools/domain"
)

// RootCommand is the parent command for all tool subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Domain domain.RootCommand `cmd:"" help:"Domain Discovery API tools"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tools"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
