package draft

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	sdkdraft "github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/draft"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to discard a routing config's draft
// version.
type DeleteCommand struct {
	argparser.Base
	routingConfigID string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Discard a routing config's draft version, reverting it back to the active version").Alias("discard")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &sdkdraft.DeleteInput{
		RoutingConfigID: &c.routingConfigID,
	}

	err := sdkdraft.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Discarded draft version for routing config (routing-config-id: %s)", c.routingConfigID)
	return nil
}
