package pop

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
}

// CommandName is the string to be used to invoke this command
const CommandName = "pops"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "List Fastly datacenters")
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
		t.AddLine(
			fastly.ToValue(dc.Name),
			fastly.ToValue(dc.Code),
			fastly.ToValue(dc.Group),
			fastly.ToValue(dc.Shield),
			Coordinates(dc.Coordinates),
		)
	}
	t.Print()
	return nil
}

// Coordinates returns a stringified object of coordinate data.
func Coordinates(c *fastly.Coordinates) string {
	if c != nil {
		return fmt.Sprintf(
			`{Latitude:%v Longtitude:%v X:%v Y:%v}`,
			fastly.ToValue(c.Latitude),
			fastly.ToValue(c.Longtitude),
			fastly.ToValue(c.X),
			fastly.ToValue(c.Y),
		)
	}
	return ""
}
