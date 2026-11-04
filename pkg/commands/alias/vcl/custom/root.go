package custom

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" hidden:"" help:"Upload a VCL for a particular service and version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete the uploaded VCL for a particular service and version"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Get the uploaded VCL for a particular service and version"`
	List     ListCommand     `cmd:"" hidden:"" help:"List the uploaded VCLs for a particular service and version"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update the uploaded VCL for a particular service and version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "custom"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
