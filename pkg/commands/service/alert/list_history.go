package alerts

import (
	"context"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// ListHistoryCommand calls the Fastly API to list appropriate resources.
type ListHistoryCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput
	After  argparser.OptionalString `help:"After filter history record that either started or ended after a specific date"`
	Before argparser.OptionalString `help:"Before filter history record that either started or ended before a specific date"`

	Cursor       argparser.OptionalString `help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	DefinitionID argparser.OptionalString `help:"Unique identifier of the definition"`
	Limit        argparser.OptionalInt    `help:"Maximum number of items to list"`
	Order        argparser.OptionalString `help:"Sort by one of the following [asc, desc]"`
	Sort         argparser.OptionalString `help:"Sort by one of the following [start]"`
	ServiceID    argparser.OptionalString `help:"ServiceID of the definition"`

	Status argparser.OptionalString `help:"Status of the history record [active, resolved]"`
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
		history, err := c.Globals.APIClient.ListAlertHistory(context.TODO(), input)
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
	if c.Cursor.WasSet {
		input.Cursor = &c.Cursor.Value
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	var sign string
	var err error
	if c.Order.WasSet {
		sign, err = argparser.ConvertOrderFromStringFlag(c.Order.Value, "order")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return nil, err
		}
	}

	if c.Sort.WasSet {
		str := sign + c.Sort.Value
		input.Sort = &str
	}

	if c.ServiceID.WasSet {
		input.ServiceID = &c.ServiceID.Value
	}

	if c.DefinitionID.WasSet {
		input.DefinitionID = &c.DefinitionID.Value
	}

	if c.Status.WasSet {
		input.Status = &c.Status.Value
	}

	if c.Before.WasSet {
		input.Before = &c.Before.Value
	}

	if c.After.WasSet {
		input.After = &c.After.Value
	}

	return &input, nil
}
