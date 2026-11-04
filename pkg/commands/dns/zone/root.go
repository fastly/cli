package zone

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a DNS Zone"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a DNS Zone"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Describe a DNS Zone"`
	List     ListCommand     `cmd:"" help:"List DNS Zones"`
	Update   UpdateCommand   `cmd:"" help:"Update a DNS Zone"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "zone"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
