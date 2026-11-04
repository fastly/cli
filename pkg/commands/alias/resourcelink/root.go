package resourcelink

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"link" hidden:"" help:"Create a Fastly service resource link"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete a resource link for a Fastly service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Show detailed information about a Fastly service resource link"`
	List     ListCommand     `cmd:"" hidden:"" help:"List all resource links for a Fastly service version"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update a resource link for a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "resource-link"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
