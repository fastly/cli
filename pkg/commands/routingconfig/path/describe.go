package path

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// DescribeCommand calls the Fastly API to describe a path.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	pathID          string
	routingConfigID string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a path").Alias("get")

	// Required.
	c.CmdClause.Flag("path-id", "The Path Identifier").Required().StringVar(&c.pathID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &paths.GetInput{
		PathID:          &c.pathID,
		RoutingConfigID: &c.routingConfigID,
	}

	d, err := paths.Get(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.pathID,
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	if d != nil {
		cl := []paths.Data{*d}
		if c.Globals.Verbose() {
			printVerbose(out, cl)
		} else {
			printSummary(out, cl)
		}
	}
	return nil
}
