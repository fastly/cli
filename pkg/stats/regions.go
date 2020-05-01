package stats

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/fastly"
)

// RegionsCommand exposes the Stats Regions API.
type RegionsCommand struct {
	common.Base
	Input  fastly.GetStatsInput
	Format string
}

// NewRegionsCommand returns a new command registered under parent.
func NewRegionsCommand(parent common.Registerer, globals *config.Data) *RegionsCommand {
	var c RegionsCommand
	c.Globals = globals
	c.CmdClause = parent.Command("regions", "List stats regions")
	return &c
}

// Exec implements the command interface.
func (c *RegionsCommand) Exec(in io.Reader, out io.Writer) error {
	resp, err := c.Globals.Client.GetRegions()
	if err != nil {
		return fmt.Errorf("regions: %w", err)
	}

	for _, region := range resp.Data {
		fmt.Fprintf(out, "%s\n", region)
	}

	return nil
}
