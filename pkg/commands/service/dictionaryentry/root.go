package dictionaryentry

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" help:"Create a new item on a Fastly edge dictionary"`
	Delete   DeleteCommand   `cmd:"" help:"Delete an item from a Fastly edge dictionary"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about a Fastly edge dictionary item"`
	List     ListCommand     `cmd:"" help:"List items in a Fastly edge dictionary"`
	Update   UpdateCommand   `cmd:"" help:"Update or insert an item on a Fastly edge dictionary"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "dictionary-entry"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
