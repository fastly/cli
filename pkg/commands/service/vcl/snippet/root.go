package snippet

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a snippet for a particular service and version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a specific snippet for a particular service and version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Get the uploaded VCL snippet for a particular service and version"`
	List     ListCommand     `cmd:"" help:"List the uploaded VCL snippets for a particular service and version"`
	Update   UpdateCommand   `cmd:"" help:"Update a VCL snippet for a particular service and version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "snippet"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
