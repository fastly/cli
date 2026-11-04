package iplist

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a workspace-level ip list"`
	Delete DeleteCommand `cmd:"" help:"Delete an account ip list"`
	Get    GetCommand    `cmd:"" help:"Get a workspace-level ip list"`
	List   ListCommand   `cmd:"" help:"List all ip lists for your workspace"`
	Update UpdateCommand `cmd:"" help:"Update a workspace-level ip list"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "ip-list"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
