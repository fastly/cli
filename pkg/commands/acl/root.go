package acl

import (
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/urfave/cli/v2"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
// type RootCommand struct {
// 	cmd.Base
// 	// no flags
// }

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(globals *config.Data, data manifest.Data) *cli.Command {
	// var c RootCommand
	c := cli.Command{
		Name:  "acl",
		Usage: "Manipulate Fastly ACLs (Access Control Lists)",
		Subcommands: []*cli.Command{
			NewCreateCommand(globals, data),
		},
		// Action: func(cCtx *cli.Context) error {
		// 	fmt.Println("added task: ", cCtx.Args().First())
		// 	return nil
		// },
	}
	// c.Globals = globals
	// c.CmdClause = parent.Command("acl", "Manipulate Fastly ACLs (Access Control Lists)")
	return &c
}

// // Exec implements the command interface.
// func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
// 	panic("unreachable")
// }
