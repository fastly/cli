package draft

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	sdkdraft "github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/draft"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to set the comment on a routing config's
// draft version.
type UpdateCommand struct {
	argparser.Base

	// Required.
	comment         string
	routingConfigID string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Set the comment on a routing config's draft version")

	// Required.
	c.CmdClause.Flag("comment", "A descriptive note about the changes made in the draft").Required().StringVar(&c.comment)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &sdkdraft.UpdateInput{
		Comment:         &c.comment,
		RoutingConfigID: &c.routingConfigID,
	}

	_, err := sdkdraft.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	text.Success(out, "Updated draft comment for routing config (routing-config-id: %s)", c.routingConfigID)
	return nil
}
