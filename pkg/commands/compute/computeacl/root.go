package computeacl

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create      CreateCommand      `cmd:"" help:"Create a compute ACL"`
	ListAcls    ListCommand        `cmd:"" help:"List all compute ACLs"`
	Describe    DescribeCommand    `cmd:"" help:"Describe a compute ACL"`
	Update      UpdateCommand      `cmd:"" help:"Update a compute ACL"`
	Lookup      LookupCommand      `cmd:"" help:"Find a matching ACL entry for an IP address"`
	Delete      DeleteCommand      `cmd:"" help:"Delete a compute ACL"`
	ListEntries ListEntriesCommand `cmd:"" help:"List all entries of a compute ACL"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "acl"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
