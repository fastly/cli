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

// UpdateCommand calls the Fastly API to set the comment on a routing config's
// draft version.
type UpdateCommand struct {
	argparser.Base

	// Required.
	Comment         string `required:"" help:"A descriptive note about the changes made in the draft"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &sdkdraft.UpdateInput{
		Comment:         &c.Comment,
		RoutingConfigID: &c.RoutingConfigID,
	}

	_, err := sdkdraft.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Updated draft comment for routing config (routing-config-id: %s)", c.RoutingConfigID)
	return nil
}
