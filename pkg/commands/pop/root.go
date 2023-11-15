package pop

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("pops", "List Fastly datacenters")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	dcs, err := c.Globals.APIClient.AllDatacenters()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Break(out)
	t := text.NewTable(out)
	t.AddHeader("NAME", "CODE", "GROUP", "SHIELD", "COORDINATES")
	for _, dc := range dcs {
		t.AddLine(dc.Name, dc.Code, dc.Group, dc.Shield, fmt.Sprintf("%+v", dc.Coordinates))
	}
	t.Print()
	return nil
}
