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

// DeleteCommand calls the Fastly API to delete a routing config.
type DeleteCommand struct {
	argparser.Base
	routingConfigID string
	force           bool
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a routing config. This is rejected if the routing config is still associated with any domain, regardless of --force; unset those associations first with 'fastly domain update --routing-config-id nil'").Alias("remove")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.CmdClause.Flag("force", "Delete the routing config even if it has an active version, bypassing the active-version check. You still can't delete a routing config that's linked to a domain").BoolVar(&c.force)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &routingconfigs.DeleteInput{
		RoutingConfigID: &c.routingConfigID,
	}
	if c.force {
		input.Force = &c.force
	}

	err := routingconfigs.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Deleted routing config (routing-config-id: %s)", c.routingConfigID)
	return nil
}
