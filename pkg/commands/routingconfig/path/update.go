package path

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a path.
type UpdateCommand struct {
	argparser.Base

	// Required.
	pathID          string
	routingConfigID string

	// Optional.
	path argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a path")

	// Required.
	c.CmdClause.Flag("path-id", "The Path Identifier").Required().StringVar(&c.pathID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.CmdClause.Flag("path", "The URL path pattern (max 2048 characters, starts with \"/\")").Action(c.path.Set).StringVar(&c.path.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &paths.UpdateInput{
		PathID:          &c.pathID,
		RoutingConfigID: &c.routingConfigID,
	}
	if c.path.WasSet {
		input.Path = &c.path.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := paths.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.pathID,
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Updated path '%s' (path-id: %s)", d.Path, d.PathID)
	return nil
}
