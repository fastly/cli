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

// CreateCommand calls the Fastly API to create a path within a routing
// config.
type CreateCommand struct {
	argparser.Base

	// Required.
	path            string
	routingConfigID string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a path within a routing config").Alias("add")

	// Required.
	c.CmdClause.Flag("path", "The URL path pattern (max 2048 characters, starts with \"/\")").Required().StringVar(&c.path)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &paths.CreateInput{
		Path:            &c.path,
		RoutingConfigID: &c.routingConfigID,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := paths.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path":              c.path,
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Created path '%s' (path-id: %s) in routing config (routing-config-id: %s)", d.Path, d.PathID, c.routingConfigID)
	return nil
}
