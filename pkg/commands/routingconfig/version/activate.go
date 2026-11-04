package version

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/versions"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// ActivateCommand calls the Fastly API to reactivate a routing config's
// inactive version.
type ActivateCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
	VersionID       string `required:"" help:"The identifier of the version to reactivate"`
}

// Exec invokes the application logic for the command.
func (c *ActivateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &versions.ActivateInput{
		RoutingConfigID: &c.RoutingConfigID,
		VersionID:       &c.VersionID,
	}

	d, err := versions.Activate(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
			"Version ID":        c.VersionID,
		})
		return err
	}

	text.Success(out, "Reactivated version '%s' for routing config '%s' (routing-config-id: %s)", c.VersionID, d.Name, d.RoutingConfigID)
	return nil
}
