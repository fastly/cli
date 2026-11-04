package path

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/paths"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a path within a routing
// config.
type CreateCommand struct {
	argparser.Base

	// Required.
	Path            string `required:"" help:"The URL path pattern (max 2048 characters, starts with \"/\")"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &paths.CreateInput{
		Path:            &c.Path,
		RoutingConfigID: &c.RoutingConfigID,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := paths.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path":              c.Path,
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Created path '%s' (path-id: %s) in routing config (routing-config-id: %s)", d.Path, d.PathID, c.RoutingConfigID)
	return nil
}
