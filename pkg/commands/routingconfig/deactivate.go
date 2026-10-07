package routingconfig

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeactivateCommand calls the Fastly API to deactivate a routing config's
// active version.
type DeactivateCommand struct {
	argparser.Base
	routingConfigID string
}

// NewDeactivateCommand returns a usable command registered under the parent.
func NewDeactivateCommand(parent argparser.Registerer, g *global.Data) *DeactivateCommand {
	c := DeactivateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("deactivate", "Deactivate a routing config's active version")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeactivateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &routingconfigs.DeactivateInput{
		RoutingConfigID: &c.routingConfigID,
	}

	d, err := routingconfigs.Deactivate(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Deactivated routing config '%s' (routing-config-id: %s)", d.Name, d.RoutingConfigID)
	return nil
}
