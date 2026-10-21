package draft

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	sdkdraft "github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/draft"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// DiffCommand calls the Fastly API to diff a routing config's draft and
// active versions.
type DiffCommand struct {
	argparser.Base
	argparser.JSONOutput
	routingConfigID string
}

// NewDiffCommand returns a usable command registered under the parent.
func NewDiffCommand(parent argparser.Registerer, g *global.Data) *DiffCommand {
	c := DiffCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("diff", "Show the differences between a routing config's active and draft versions")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
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
		RoutingConfigID: &c.routingConfigID,
	}

	d, err := sdkdraft.GetDiff(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
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
