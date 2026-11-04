package auditlog

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/auditlog/eventmapping"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	EventMapping eventmapping.RootCommand `cmd:"" help:"Manage Fastly audit log event mappings"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "audit-log"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
