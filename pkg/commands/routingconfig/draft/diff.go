package draft

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	sdkdraft "github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/draft"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DiffCommand calls the Fastly API to diff a routing config's draft and
// active versions.
type DiffCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DiffCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &sdkdraft.GetDiffInput{
		RoutingConfigID: &c.RoutingConfigID,
	}

	d, err := sdkdraft.GetDiff(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	for _, added := range d.Added {
		if added.Path != nil {
			fmt.Fprintf(out, "Added path '%s' (path-id: %s)\n", added.Path.Path, added.Path.PathID)
		}
	}
	for _, deleted := range d.Deleted {
		if deleted.Path != nil {
			fmt.Fprintf(out, "Deleted path '%s' (path-id: %s)\n", deleted.Path.Path, deleted.Path.PathID)
		}
	}
	for _, modified := range d.Modified {
		fmt.Fprintf(out, "Modified path (path-id: %s): %d rule(s) added, %d rule(s) changed, %d rule(s) deleted\n",
			modified.PathID, len(modified.RulesAdded), len(modified.RulesChanged), len(modified.RulesDeleted))
	}
	return nil
}
