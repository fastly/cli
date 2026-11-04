package draft

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	sdkdraft "github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/draft"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to discard a routing config's draft
// version.
type DeleteCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &sdkdraft.DeleteInput{
		RoutingConfigID: &c.RoutingConfigID,
	}

	err := sdkdraft.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Discarded draft version for routing config (routing-config-id: %s)", c.RoutingConfigID)
	return nil
}
