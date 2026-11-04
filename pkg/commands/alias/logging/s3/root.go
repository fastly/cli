package s3

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" hidden:"" help:"Create an Amazon S3 logging endpoint on a Fastly service version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete a S3 logging endpoint on a Fastly service version"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Show detailed information about a S3 logging endpoint on a Fastly service version"`
	List     ListCommand     `cmd:"" hidden:"" help:"List S3 endpoints on a Fastly service version"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update a S3 logging endpoint on a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "s3"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
