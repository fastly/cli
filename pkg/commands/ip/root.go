package ip

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
	c.CmdClause = parent.Command("ip-list", "List Fastly's public IPs")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	ipv4, ipv6, err := c.Globals.APIClient.AllIPs()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	// TODO: Implement --json support.

	text.Break(out)
	fmt.Fprintf(out, "%s\n", text.Bold("IPv4"))
	for _, ip := range ipv4 {
		fmt.Fprintf(out, "\t%s\n", ip)
	}
	fmt.Fprintf(out, "\n%s\n", text.Bold("IPv6"))
	for _, ip := range ipv6 {
		fmt.Fprintf(out, "\t%s\n", ip)
	}
	return nil
}
