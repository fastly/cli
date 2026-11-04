package apisecurity

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/apisecurity/discoveredoperations"
	"github.com/fastly/cli/pkg/commands/apisecurity/operations"
	"github.com/fastly/cli/pkg/commands/apisecurity/tags"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	DiscoveredOperations discoveredoperations.RootCommand `cmd:"" help:"Retrieve and update discovered API operations"`
	Operations           operations.RootCommand           `cmd:"" help:"Manage operations associated with services"`
	Tags                 tags.RootCommand                 `cmd:"" help:"Manipulate Fastly API Security operation tags"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "apisecurity"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
