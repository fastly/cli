package routingconfig

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeactivateCommand calls the Fastly API to deactivate a routing config's
// active version.
type DeactivateCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *DeactivateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &routingconfigs.DeactivateInput{
		RoutingConfigID: &c.RoutingConfigID,
	}

	d, err := routingconfigs.Deactivate(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Deactivated routing config '%s' (routing-config-id: %s)", d.Name, d.RoutingConfigID)
	text.Warning(out, "This does not unlink the routing config from any domains. If domains are still associated with it, unset them with 'fastly domain update --domain-id <id> --routing-config-id nil'.")
	return nil
}
