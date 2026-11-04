package alerts

import (
	"context"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput

	Cursor         argparser.OptionalString `help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	Limit          argparser.OptionalInt    `help:"Maximum number of items to list"`
	DefinitionName argparser.OptionalString `name:"name" help:"Name of the definition"`
	Order          argparser.OptionalString `help:"Sort by one of the following [asc, desc]"`
	Sort           argparser.OptionalString `help:"Sort by one of the following [name, created_at, updated_at]"`
	ServiceID      argparser.OptionalString `help:"ServiceID of the definition"`
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
		definitions, err := c.Globals.APIClient.ListAlertDefinitions(context.TODO(), input)
		if err != nil {
			return err
		}

		if ok, err := c.WriteJSON(out, definitions); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		definitionsPtr := make([]*fastly.AlertDefinition, len(definitions.Data))
		for i := range definitions.Data {
			definitionsPtr[i] = &definitions.Data[i]
		}

		if c.Globals.Verbose() {
			printVerbose(out, definitionsPtr)
		} else {
			printSummary(out, definitionsPtr)
		}

		if definitions != nil && definitions.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &definitions.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() (*fastly.ListAlertDefinitionsInput, error) {
	input := fastly.ListAlertDefinitionsInput{}
	if c.Cursor.WasSet {
		input.Cursor = &c.Cursor.Value
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	if c.DefinitionName.WasSet {
		input.Name = &c.DefinitionName.Value
	}
	if c.ServiceID.WasSet {
		input.ServiceID = &c.ServiceID.Value
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
