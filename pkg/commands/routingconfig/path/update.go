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

// UpdateCommand calls the Fastly API to update a path.
type UpdateCommand struct {
	argparser.Base

	// Required.
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	// Optional.
	Path argparser.OptionalString `help:"The URL path pattern (max 2048 characters, starts with \"/\")"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &paths.UpdateInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
	}
	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := paths.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Updated path '%s' (path-id: %s)", d.Path, d.PathID)
	return nil
}
