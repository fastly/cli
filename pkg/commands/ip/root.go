package ip

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	json bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("ip-list", "List Fastly's public IPs")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	ipv4, ipv6, err := c.Globals.APIClient.AllIPs()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		type IPs struct {
			IPv4 fastly.IPAddrs
			IPv6 fastly.IPAddrs
		}
		ips := IPs{ipv4, ipv6}
		data, err := json.Marshal(&ips)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
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
