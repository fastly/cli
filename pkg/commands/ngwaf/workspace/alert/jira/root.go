package jira

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a Jira alert"`
	Delete DeleteCommand `cmd:"" help:"Delete a Jira alert"`
	Get    GetCommand    `cmd:"" help:"Get a Jira alert"`
	List   ListCommand   `cmd:"" help:"List Jira alerts"`
	Update UpdateCommand `cmd:"" help:"Update a Jira alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "jira"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// ConfigFlags contains Jira specific configuration flags.
type ConfigFlags struct {
	Host     string
	Key      string
	Project  string
	Username string
}

// OptConfigFlags contains optional Jira specific configuration flags.
type OptConfigFlags struct {
	IssueType string
}
