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

// ListCommand calls the Fastly API to list paths within a routing config.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	routingConfigID string

	// Optional.
	match argparser.OptionalString
	path  argparser.OptionalString
	sort  argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List paths within a routing config")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("match", "Filters results using the given path pattern matching strategy").Action(c.match.Set).StringVar(&c.match.Value)
	c.CmdClause.Flag("path", "Filters results by path pattern").Action(c.path.Set).StringVar(&c.path.Value)
	c.CmdClause.Flag("sort", "The order in which to list the results").Action(c.sort.Set).StringVar(&c.sort.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &paths.ListInput{
		RoutingConfigID: &c.routingConfigID,
	}
	if c.match.WasSet {
		input.Match = &c.match.Value
	}
	if c.path.WasSet {
		input.Path = &c.path.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := paths.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
