package version

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/versions"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteInactiveCommand calls the Fastly API to permanently delete all
// inactive versions of a routing config.
type DeleteInactiveCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *DeleteInactiveCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &versions.DeleteInactiveInput{
		RoutingConfigID: &c.RoutingConfigID,
	}

	err := versions.DeleteInactive(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Deleted inactive versions for routing config (routing-config-id: %s)", c.RoutingConfigID)
	return nil
}
