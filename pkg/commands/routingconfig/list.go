package routingconfig

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// ListCommand calls the Fastly API to list routing configs.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	limit argparser.OptionalInt
	sort  argparser.OptionalString
	state argparser.OptionalStringSlice
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List routing configs")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("limit", "Limit how many results are returned per page").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("sort", "The order in which to list the results").Action(c.sort.Set).StringVar(&c.sort.Value)
	c.CmdClause.Flag("state", "Filter results by lifecycle state. Set flag multiple times to filter by multiple states").Action(c.state.Set).StringsVar(&c.state.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &routingconfigs.ListInput{}

	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}
	if c.state.WasSet {
		input.State = c.state.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := routingconfigs.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Limit": c.limit.Value,
			"Sort":  c.sort.Value,
			"State": c.state.Value,
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
