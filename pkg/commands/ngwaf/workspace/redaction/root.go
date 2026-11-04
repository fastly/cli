package redaction

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand `cmd:"" aliases:"add" help:"Create a redaction"`
	Delete   DeleteCommand `cmd:"" help:"Delete a redaction"`
	List     ListCommand   `cmd:"" help:"List redactions in a workspace"`
	Retrieve GetCommand    `cmd:"" aliases:"get" help:"Retrieve a redaction"`
	Update   UpdateCommand `cmd:"" help:"Update a redaction"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "redaction"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
