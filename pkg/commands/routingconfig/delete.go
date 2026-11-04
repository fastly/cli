package routingconfig

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a routing config.
type DeleteCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	// Optional.
	Force bool `help:"Delete the routing config even if it has an active version, bypassing the active-version check. You still can't delete a routing config that's linked to a domain"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &routingconfigs.DeleteInput{
		RoutingConfigID: &c.RoutingConfigID,
	}
	if c.Force {
		input.Force = &c.Force
	}

	err := routingconfigs.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Deleted routing config (routing-config-id: %s)", c.RoutingConfigID)
	return nil
}
