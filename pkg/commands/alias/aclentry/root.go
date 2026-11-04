package aclentry

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" hidden:"" help:"Add an ACL entry to an ACL"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete an ACL entry from a specified ACL"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Retrieve a single ACL entry"`
	List     ListCommand     `cmd:"" hidden:"" help:"List ACLs"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update an ACL entry for a specified ACL"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "aclentry"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
