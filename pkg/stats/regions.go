package stats

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// RegionsCommand exposes the Stats Regions API.
type RegionsCommand struct {
	common.Base
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
		return fmt.Errorf("fetching regions: %w", err)
	}

	for _, region := range resp.Data {
		text.Output(out, "%s", region)
	}

	return nil
}
