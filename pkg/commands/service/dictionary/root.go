package dictionary

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" help:"Create a Fastly edge dictionary on a Fastly service version"`
	Delete   DeleteCommand   `cmd:"" help:"Delete a Fastly edge dictionary from a Fastly service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about a Fastly edge dictionary"`
	List     ListCommand     `cmd:"" help:"List all dictionaries on a Fastly service version"`
	Update   UpdateCommand   `cmd:"" aliases:"get" help:"Update name of dictionary on a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "dictionary"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
