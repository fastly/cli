package vcl

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/alias/vcl/condition"
	"github.com/fastly/cli/pkg/commands/alias/vcl/custom"
	"github.com/fastly/cli/pkg/commands/alias/vcl/snippet"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Describe  DescribeCommand       `cmd:"" aliases:"get" hidden:"" help:"Get the generated VCL for a particular service and version"`
	Condition condition.RootCommand `cmd:"" hidden:"" help:"Manipulate Fastly VCL conditions"`
	Custom    custom.RootCommand    `cmd:"" hidden:"" help:"Manipulate Fastly custom VCL files"`
	Snippet   snippet.RootCommand   `cmd:"" hidden:"" help:"Manipulate Fastly VCL snippets"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "vcl"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
