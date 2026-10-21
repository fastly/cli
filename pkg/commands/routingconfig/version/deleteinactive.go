package version

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/versions"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteInactiveCommand calls the Fastly API to permanently delete all
// inactive versions of a routing config.
type DeleteInactiveCommand struct {
	argparser.Base
	routingConfigID string
}

// NewDeleteInactiveCommand returns a usable command registered under the
// parent.
func NewDeleteInactiveCommand(parent argparser.Registerer, g *global.Data) *DeleteInactiveCommand {
	c := DeleteInactiveCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete-inactive", "Permanently delete all inactive versions of a routing config. This destroys rollback history and cannot be undone")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteInactiveCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &versions.DeleteInactiveInput{
		RoutingConfigID: &c.routingConfigID,
	}

	err := versions.DeleteInactive(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Deleted inactive versions for routing config (routing-config-id: %s)", c.routingConfigID)
	return nil
}
