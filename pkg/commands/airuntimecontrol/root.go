package airuntimecontrol

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol/key"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol/provider"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol/providerconnection"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol/session"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol/usagemetrics"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Key                key.RootCommand                `cmd:"" help:"Manage AI Runtime Control virtual keys"`
	Provider           provider.RootCommand           `cmd:"" help:"Manage AI Runtime Control providers"`
	ProviderConnection providerconnection.RootCommand `cmd:"" help:"Manage AI Runtime Control provider connections"`
	Session            session.RootCommand            `cmd:"" help:"Manage AI Runtime Control session logs"`
	UsageMetrics       usagemetrics.RootCommand       `cmd:"" help:"Manage AI Runtime Control usage metrics"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "ai-runtime-control"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
