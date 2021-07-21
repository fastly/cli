package ip

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("ip-list", "List Fastly's public IPs")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	ipv4, ipv6, err := c.Globals.Client.AllIPs()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

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
