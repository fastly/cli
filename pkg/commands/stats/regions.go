package stats

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// RegionsCommand exposes the Stats Regions API.
type RegionsCommand struct {
	argparser.Base
}

// Exec implements the command interface.
func (c *RegionsCommand) Exec(_ io.Reader, out io.Writer) error {
	resp, err := c.Globals.APIClient.GetRegions(context.TODO())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("fetching regions: %w", err)
	}

	for _, region := range resp.Data {
		text.Output(out, "%s", region)
	}

	return nil
}
