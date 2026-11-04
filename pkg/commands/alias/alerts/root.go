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

	Create   CreateCommand      `cmd:"" hidden:"" help:"Create Alert"`
	Delete   DeleteCommand      `cmd:"" hidden:"" help:"Delete Alert"`
	Describe DescribeCommand    `cmd:"" hidden:"" help:"Describe Alert"`
	List     ListCommand        `cmd:"" hidden:"" help:"List Alerts"`
	History  ListHistoryCommand `cmd:"" hidden:"" help:"List history"`
	Update   UpdateCommand      `cmd:"" hidden:"" help:"Update Alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "alerts"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
