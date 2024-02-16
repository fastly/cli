package resourcelink

import (
	"io"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/global"
)

// RootName is the name of this package's sub-command in the CLI,
// e.g. "fastly resource-link".
const RootName = "resource-link"

// Common flag descriptions.
const (
	flagNameDescription       = "Resource link name (alias). Defaults to resource's name"
	flagIDDescription         = "Resource link ID"
	flagResourceIDDescription = "Resource ID"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(RootName, "Manipulate Fastly service resource links")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
