package rule

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a rule within a path"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about a rule"`
	List     ListCommand     `cmd:"" help:"List rules within a path"`
	Update   UpdateCommand   `cmd:"" help:"Update a rule"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a rule"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "rule"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
