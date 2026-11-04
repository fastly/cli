package path

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/routingconfig/path/rule"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand    `cmd:"" aliases:"add" help:"Create a path within a routing config"`
	Describe DescribeCommand  `cmd:"" aliases:"get" help:"Show detailed information about a path"`
	List     ListCommand      `cmd:"" help:"List paths within a routing config"`
	Update   UpdateCommand    `cmd:"" help:"Update a path"`
	Delete   DeleteCommand    `cmd:"" aliases:"remove" help:"Delete a path"`
	Rule     rule.RootCommand `cmd:"" help:"Manipulate rules within a path of a Fastly routing config"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "path"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
