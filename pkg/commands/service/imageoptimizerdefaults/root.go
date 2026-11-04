package imageoptimizerdefaults

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Get    GetCommand    `cmd:"" help:"Retrieve the current Image Optimizer default settings"`
	Update UpdateCommand `cmd:"" help:"Update Image Optimizer default settings for a service"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "imageoptimizer"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
