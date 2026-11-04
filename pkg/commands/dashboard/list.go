package dashboard

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/printer"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional Flags
	argparser.JSONOutput

	Cursor argparser.OptionalString `help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	Limit  argparser.OptionalInt    `help:"Maximum number of items to list"`
	Order  argparser.OptionalString `help:"Sort by one of the following [asc, desc]"`
	Sort   argparser.OptionalString `help:"Sort by one of the following [name, created_at, updated_at]"`
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

	var dashboards []fastly.ObservabilityCustomDashboard
	loadAllPages := c.JSONOutput.Enabled || c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes

	for {
		o, err := c.Globals.APIClient.ListObservabilityCustomDashboards(context.TODO(), input)
		if err != nil {
			return err
		}

		if o != nil {
			dashboards = append(dashboards, o.Data...)

			if loadAllPages {
				if o.Meta.NextCursor != "" {
					input.Cursor = &o.Meta.NextCursor
					continue
				}
				break
			}

			if c.Globals.Verbose() {
				printer.PrintVerbose(out, dashboards)
			} else {
				printer.PrintSummary(out, dashboards)
			}

			if o.Meta.NextCursor != "" && text.IsTTY(out) {
				text.Break(out)
				printNextPage, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNextPage {
					dashboards = []fastly.ObservabilityCustomDashboard{}
					input.Cursor = &o.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}

	if ok, err := c.WriteJSON(out, dashboards); ok {
		// No pagination prompt w/ JSON output.
		return err
	}

	// Only print output here if we've not already printed JSON.
	if c.Globals.Verbose() {
		printer.PrintVerbose(out, dashboards)
	} else {
		printer.PrintSummary(out, dashboards)
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() (*fastly.ListObservabilityCustomDashboardsInput, error) {
	var input fastly.ListObservabilityCustomDashboardsInput

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

	return &input, nil
}
