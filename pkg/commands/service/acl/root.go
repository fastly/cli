package acl

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a new ACL attached to the specified service version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete an ACL from the specified service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single ACL by name for the version and service"`
	List     ListCommand     `cmd:"" help:"List ACLs"`
	Update   UpdateCommand   `cmd:"" help:"Update an ACL for a particular service and version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "acl"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
