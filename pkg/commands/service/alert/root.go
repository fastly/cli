package alerts

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand      `cmd:"" help:"Create Alert"`
	Delete   DeleteCommand      `cmd:"" help:"Delete Alert"`
	Describe DescribeCommand    `cmd:"" help:"Describe Alert"`
	List     ListCommand        `cmd:"" help:"List Alerts"`
	History  ListHistoryCommand `cmd:"" help:"List history"`
	Update   UpdateCommand      `cmd:"" help:"Update Alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "alert"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
