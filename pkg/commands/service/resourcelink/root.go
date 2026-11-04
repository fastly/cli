package resourcelink

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
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

	Create   CreateCommand   `cmd:"" aliases:"link" help:"Create a Fastly service resource link"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a resource link for a Fastly service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about a Fastly service resource link"`
	List     ListCommand     `cmd:"" help:"List all resource links for a Fastly service version"`
	Update   UpdateCommand   `cmd:"" help:"Update a resource link for a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "resource-link"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
