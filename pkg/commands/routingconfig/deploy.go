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

// DeployCommand calls the Fastly API to activate a routing config's draft
// version.
type DeployCommand struct {
	argparser.Base
	routingConfigID string
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent argparser.Registerer, g *global.Data) *DeployCommand {
	c := DeployCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("deploy", "Activate the draft version of a routing config").Alias("activate")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeployCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &routingconfigs.ActivateInput{
		RoutingConfigID: &c.routingConfigID,
	}

	d, err := routingconfigs.Activate(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Activated routing config '%s' (routing-config-id: %s)", d.Name, d.RoutingConfigID)
	return nil
}
