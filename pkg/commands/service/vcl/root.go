package vcl

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/vcl/condition"
	"github.com/fastly/cli/pkg/commands/service/vcl/custom"
	"github.com/fastly/cli/pkg/commands/service/vcl/snippet"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Describe  DescribeCommand       `cmd:"" aliases:"get" help:"Get the generated VCL for a particular service and version"`
	Condition condition.RootCommand `cmd:"" help:"Manipulate Fastly service version conditions"`
	Custom    custom.RootCommand    `cmd:"" help:"Manipulate Fastly service version custom VCL"`
	Snippet   snippet.RootCommand   `cmd:"" help:"Manipulate Fastly VCL snippets (blocks of VCL logic inserted into your service's configuration that don't require custom VCL)"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "vcl"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
