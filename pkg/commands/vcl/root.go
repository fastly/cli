package vcl

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	// no flags
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("vcl", "Manipulate Fastly service version VCL")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	panic("unreachable")
}
