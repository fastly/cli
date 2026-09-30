package category

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/argparser/validate"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/kingpin"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// CommandName is the string to be used to invoke this command.
const CommandName = "category"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Manage Bot Management categories within a workspace")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

type categoryFilterFields struct {
	// Required.
	categoryID  string
	workspaceID string
}

// register adds the flags required to identify a bot for the API.
func (f *categoryFilterFields) register(clause *kingpin.CmdClause) {
	// Though go-fastly validates that the ids are not nil or empty,
	// we get more consistent error messages by checking length up front.
	// Otherwise there's nothing linking an API field back to the CLI
	// flag that was used to set it.

	clause.Flag("category-id", "Category ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&f.categoryID)

	clause.Flag("workspace-id", "Workspace ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&f.workspaceID)
}
