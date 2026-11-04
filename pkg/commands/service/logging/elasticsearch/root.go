package elasticsearch

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create an Elasticsearch logging endpoint on a Fastly service version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete an Elasticsearch logging endpoint on a Fastly service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about an Elasticsearch logging endpoint on a Fastly service version"`
	List     ListCommand     `cmd:"" help:"List Elasticsearch endpoints on a Fastly service version"`
	Update   UpdateCommand   `cmd:"" help:"Update an Elasticsearch logging endpoint on a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "elasticsearch"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
