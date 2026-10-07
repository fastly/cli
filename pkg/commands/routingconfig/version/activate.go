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

// ActivateCommand calls the Fastly API to reactivate a routing config's
// inactive version.
type ActivateCommand struct {
	argparser.Base

	// Required.
	routingConfigID string
	versionID       string
}

// NewActivateCommand returns a usable command registered under the parent.
func NewActivateCommand(parent argparser.Registerer, g *global.Data) *ActivateCommand {
	c := ActivateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("activate", "Reactivate a previous version of a routing config")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)
	c.CmdClause.Flag("version-id", "The identifier of the version to reactivate").Required().StringVar(&c.versionID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *ActivateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &versions.ActivateInput{
		RoutingConfigID: &c.routingConfigID,
		VersionID:       &c.versionID,
	}

	d, err := versions.Activate(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
			"Version ID":        c.versionID,
		})
		return err
	}

	text.Success(out, "Activated version '%s' for routing config '%s' (routing-config-id: %s)", c.versionID, d.Name, d.RoutingConfigID)
	return nil
}
