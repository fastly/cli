package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// ListCommand calls the Fastly API to list rules within a path.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	pathID          string
	routingConfigID string

	// Optional.
	limit argparser.OptionalInt
	sort  argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List rules within a path")

	// Required.
	c.CmdClause.Flag("path-id", "The Path Identifier").Required().StringVar(&c.pathID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("limit", "Limit how many results are returned per page").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("sort", "The order in which to list the results").Action(c.sort.Set).StringVar(&c.sort.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &rules.ListInput{
		PathID:          &c.pathID,
		RoutingConfigID: &c.routingConfigID,
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := rules.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.pathID,
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, cl); ok {
		return err
	}

	if c.Globals.Verbose() {
		printVerbose(out, cl)
	} else {
		printSummary(out, cl)
	}
	return nil
}
