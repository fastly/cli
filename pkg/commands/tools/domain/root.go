package domain

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all tool subcommands in this package.
// It should be installed under the primary `tools` root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Status  GetDomainStatusCommand      `cmd:"" help:"Check domain name availability"`
	Suggest GetDomainSuggestionsCommand `cmd:"" help:"Request domain search results for a given query"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "domain"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
