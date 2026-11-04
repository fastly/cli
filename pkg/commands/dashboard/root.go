package dashboard

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/item"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List     ListCommand      `cmd:"" help:"List custom dashboards"`
	Create   CreateCommand    `cmd:"" aliases:"add" help:"Create a custom dashboard"`
	Describe DescribeCommand  `cmd:"" aliases:"get" help:"Show detailed information about a custom dashboard"`
	Update   UpdateCommand    `cmd:"" help:"Update a custom dashboard"`
	Delete   DeleteCommand    `cmd:"" aliases:"remove" help:"Delete a custom dashboard"`
	Item     item.RootCommand `cmd:"" help:"Manipulate Fastly Custom Dashboard Items"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "dashboard"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
