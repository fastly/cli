package kvstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"insert" help:"Insert a key-value pair"`
	Delete   DeleteCommand   `cmd:"" help:"Delete one, multiple, or all keys"`
	Get      GetCommand      `cmd:"" help:"Get the value associated with a key"`
	Describe DescribeCommand `cmd:"" help:"Get the associated attributes of a key"`
	List     ListCommand     `cmd:"" help:"List keys"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "kv-store-entry"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
