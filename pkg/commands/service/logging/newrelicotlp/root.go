package newrelicotlp

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create an New Relic logging endpoint attached to the specified service version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete the New Relic OTLP Logs logging object for a particular service and version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Get the details of a New Relic OTLP Logs logging object for a particular service and version"`
	List     ListCommand     `cmd:"" help:"List all of the New Relic OTLP Logs logging objects for a particular service and version"`
	Update   UpdateCommand   `cmd:"" help:"Update a New Relic Logs logging object for a particular service and version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "newrelicotlp"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
