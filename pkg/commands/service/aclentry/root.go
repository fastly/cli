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

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Add an ACL entry to an ACL"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete an ACL entry from a specified ACL"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single ACL entry"`
	List     ListCommand     `cmd:"" help:"List ACLs"`
	Update   UpdateCommand   `cmd:"" help:"Update an ACL entry for a specified ACL"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "acl-entry"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
