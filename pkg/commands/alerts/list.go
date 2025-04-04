package alerts

import (
	"errors"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Alerts")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("cursor", "Pagination cursor (Use 'next_cursor' value from list output)").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.CmdClause.Flag("limit", "Maximum number of items to list").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("name", "Name of the definition").Action(c.definitionName.Set).StringVar(&c.definitionName.Value)
	c.CmdClause.Flag("order", "Sort by one of the following [asc, desc]").Action(c.order.Set).StringVar(&c.order.Value)
	c.CmdClause.Flag("sort", "Sort by one of the following [name, created_at, updated_at]").Action(c.sort.Set).StringVar(&c.sort.Value)
	c.CmdClause.Flag(argparser.FlagServiceIDName, "ServiceID of the definition").Action(c.serviceID.Set).StringVar(&c.serviceID.Value) // --service-id

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	cursor         argparser.OptionalString
	limit          argparser.OptionalInt
	definitionName argparser.OptionalString
	serviceID      argparser.OptionalString
	sort           argparser.OptionalString
	order          argparser.OptionalString
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
		definitions, err := c.Globals.APIClient.ListAlertDefinitions(input)
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
	if c.cursor.WasSet {
		input.Cursor = &c.cursor.Value
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.definitionName.WasSet {
		input.Name = &c.definitionName.Value
	}
	if c.serviceID.WasSet {
		input.ServiceID = &c.serviceID.Value
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
