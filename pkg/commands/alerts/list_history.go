package alerts

import (
	"errors"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewListHistoryCommand returns a usable command registered under the parent.
func NewListHistoryCommand(parent argparser.Registerer, g *global.Data) *ListHistoryCommand {
	c := ListHistoryCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("history", "List history")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("after", "After filter history record that either started or ended after a specific date").Action(c.after.Set).StringVar(&c.after.Value)
	c.CmdClause.Flag("before", "Before filter history record that either started or ended before a specific date").Action(c.before.Set).StringVar(&c.before.Value)
	c.CmdClause.Flag("cursor", "Pagination cursor (Use 'next_cursor' value from list output)").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.CmdClause.Flag("definition-id", "Unique identifier of the definition").Action(c.definitionID.Set).StringVar(&c.definitionID.Value)
	c.CmdClause.Flag("limit", "Maximum number of items to list").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("order", "Sort by one of the following [asc, desc]").Action(c.order.Set).StringVar(&c.order.Value)
	c.CmdClause.Flag("sort", "Sort by one of the following [start]").Action(c.sort.Set).StringVar(&c.sort.Value)
	c.CmdClause.Flag(argparser.FlagServiceIDName, "ServiceID of the definition").Action(c.serviceID.Set).StringVar(&c.serviceID.Value) // --service-id
	c.CmdClause.Flag("status", "Status of the history record [active, resolved]").Action(c.status.Set).StringVar(&c.status.Value)

	return &c
}

// ListHistoryCommand calls the Fastly API to list appropriate resources.
type ListHistoryCommand struct {
	argparser.Base
	argparser.JSONOutput

	cursor argparser.OptionalString
	limit  argparser.OptionalInt
	sort   argparser.OptionalString
	order  argparser.OptionalString

	status       argparser.OptionalString
	before       argparser.OptionalString
	after        argparser.OptionalString
	definitionID argparser.OptionalString
	serviceID    argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *ListHistoryCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	for {
		history, err := c.Globals.APIClient.ListAlertHistory(input)
		if err != nil {
			return err
		}

		if ok, err := c.WriteJSON(out, history); ok {
			return err
		}

		historyPtr := make([]*fastly.AlertHistory, len(history.Data))
		for i := range history.Data {
			historyPtr[i] = &history.Data[i]
		}

		if c.Globals.Verbose() {
			printHistoryVerbose(out, historyPtr)
		} else {
			printHistorySummary(out, historyPtr)
		}

		if history != nil && history.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &history.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListHistoryCommand) constructInput() (*fastly.ListAlertHistoryInput, error) {
	input := fastly.ListAlertHistoryInput{}
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

	if c.serviceID.WasSet {
		input.ServiceID = &c.serviceID.Value
	}

	if c.definitionID.WasSet {
		input.DefinitionID = &c.definitionID.Value
	}

	if c.status.WasSet {
		input.Status = &c.status.Value
	}

	if c.before.WasSet {
		input.Before = &c.before.Value
	}

	if c.after.WasSet {
		input.After = &c.after.Value
	}

	return &input, nil
}
