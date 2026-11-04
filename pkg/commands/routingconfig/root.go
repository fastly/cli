package routingconfig

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/routingconfig/draft"
	"github.com/fastly/cli/pkg/commands/routingconfig/path"
	"github.com/fastly/cli/pkg/commands/routingconfig/version"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create     CreateCommand       `cmd:"" aliases:"add" help:"Create a routing config"`
	Describe   DescribeCommand     `cmd:"" aliases:"get" help:"Show detailed information about a routing config"`
	List       ListCommand         `cmd:"" help:"List routing configs"`
	Delete     DeleteCommand       `cmd:"" aliases:"remove" help:"Delete a routing config. This is rejected if the routing config is still associated with any domain, regardless of --force; unset those associations first with 'fastly domain update --routing-config-id nil'"`
	Deploy     DeployCommand       `cmd:"" aliases:"activate" help:"Activate the draft version of a routing config"`
	Deactivate DeactivateCommand   `cmd:"" help:"Deactivate a routing config, leaving it without an active version. This does not unlink the routing config from any domains it's associated with; unset those manually with 'fastly domain update --routing-config-id nil'"`
	Draft      draft.RootCommand   `cmd:"" help:"Manipulate a Fastly routing config's draft version"`
	Version    version.RootCommand `cmd:"" help:"Manipulate a Fastly routing config's historical versions"`
	Path       path.RootCommand    `cmd:"" help:"Manipulate paths within a Fastly routing config"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "routing-config"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
