package compute

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/compute/computeacl"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	ACL          computeacl.RootCommand `cmd:"" help:"Manipulate Fastly compute ACLs"`
	Build        BuildCommand           `cmd:"" help:"Build a Compute package locally"`
	Deploy       DeployCommand          `cmd:"" help:"Deploy a package to a Fastly Compute service"`
	HashFiles    HashFilesCommand       `cmd:"" help:"Generate a SHA512 digest from the contents of the Compute package"`
	Init         InitCommand            `cmd:"" help:"Initialize a new Compute package locally"`
	InstallTools InstallCommand         "cmd:\"\" help:\"Download and install the tools used by `compute` subcommands (currently the Viceroy binary used by `compute serve`)\""
	Metadata     MetadataCommand        `cmd:"" help:"Control what metadata is collected"`
	Pack         PackCommand            `cmd:"" help:"Package a pre-compiled Wasm binary for a Fastly Compute service"`
	Publish      PublishCommand         `cmd:"" help:"Build and deploy a Compute package to a Fastly service"`
	Serve        ServeCommand           `cmd:"" help:"Build and run a Compute package locally"`
	Update       UpdateCommand          `cmd:"" help:"Update a package on a Fastly Compute service version"`
	Validate     ValidateCommand        `cmd:"" help:"Validate a Compute package"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "compute"

// NewRootCommand returns a RootCommand set up as it was before parsing.
// Its flags now come from the struct's tags.
func NewRootCommand(g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
