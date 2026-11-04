package eventmapping

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create         CreateCommand         `cmd:"" aliases:"add" help:"Create an audit log event mapping"`
	List           ListCommand           `cmd:"" help:"List audit log event mappings"`
	Describe       DescribeCommand       `cmd:"" aliases:"get" help:"Retrieve a single audit log event mapping"`
	Update         UpdateCommand         `cmd:"" help:"Update an audit log event mapping. This replaces the entire mapping, so all required fields must be provided"`
	Delete         DeleteCommand         `cmd:"" aliases:"remove" help:"Delete an audit log event mapping"`
	ListEventTypes ListEventTypesCommand `cmd:"" help:"List the audit event types supported when creating an event mapping"`
	ListScopeTypes ListScopeTypesCommand `cmd:"" help:"List the scope types supported when creating an event mapping"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "event-mapping"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
