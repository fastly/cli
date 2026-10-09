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

// DeleteCommand calls the Fastly API to delete a path.
type DeleteCommand struct {
	argparser.Base

	// Required.
	pathID          string
	routingConfigID string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a path").Alias("remove")

	// Required.
	c.CmdClause.Flag("path-id", "The Path Identifier").Required().StringVar(&c.pathID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &paths.DeleteInput{
		PathID:          &c.pathID,
		RoutingConfigID: &c.routingConfigID,
	}

	err := paths.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.pathID,
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Deleted path (path-id: %s)", c.pathID)
	return nil
}
