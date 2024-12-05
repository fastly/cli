package dashboard

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/common"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, globals *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List custom dashboards")
	c.Globals = globals

	// Optional Flags
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("cursor", "Pagination cursor (Use 'next_cursor' value from list output)").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.CmdClause.Flag("limit", "Maximum number of items to list").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("order", "Sort by one of the following [asc, desc]").Action(c.order.Set).StringVar(&c.order.Value)
	c.CmdClause.Flag("sort", "Sort by one of the following [name, created_at, updated_at]").Action(c.sort.Set).StringVar(&c.sort.Value)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	cursor argparser.OptionalString
	limit  argparser.OptionalInt
	sort   argparser.OptionalString
	order  argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	for {
		dashboards, err := c.Globals.APIClient.ListObservabilityCustomDashboards(input)
		if err != nil {
			return err
		}

		if ok, err := c.WriteJSON(out, dashboards); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		dashboardsPtr := make([]*fastly.ObservabilityCustomDashboard, len(dashboards.Data))
		for i := range dashboards.Data {
			dashboardsPtr[i] = &dashboards.Data[i]
		}

		if c.Globals.Verbose() {
			common.PrintVerbose(out, dashboardsPtr)
		} else {
			common.PrintSummary(out, dashboardsPtr)
		}

		if dashboards != nil && dashboards.Meta.NextCursor != "" {
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &dashboards.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() (*fastly.ListObservabilityCustomDashboardsInput, error) {
	var input fastly.ListObservabilityCustomDashboardsInput

	if c.cursor.WasSet {
		input.Cursor = &c.cursor.Value
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	var sign string
	if c.order.WasSet {
		switch c.order.Value {
		case "asc":
		case "desc":
			sign = "-"
		default:
			err := errors.New("'order' flag must be one of the following [asc, desc]")
			c.Globals.ErrLog.Add(err)
			return nil, err
		}
	}

	if c.sort.WasSet {
		str := sign + c.sort.Value
		input.Sort = &str
	}

	return &input, nil
}
