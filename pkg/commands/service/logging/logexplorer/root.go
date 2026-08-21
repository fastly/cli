package logexplorer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CommandName is the string used to invoke the Log Explorer command.
const CommandName = "log-explorer"

// logExplorerFilterFieldOptions is a string representation of the filter
// fields supported by go-fastly.
var logExplorerFilterFieldOptions = func() (fields []string) {
	for _, field := range fastly.LogExplorerFilterFields {
		fields = append(fields, string(field))
	}
	return fields
}()

// logExplorerFilterOperatorOptions is a string representation of the filter
// operators supported by go-fastly.
var logExplorerFilterOperatorOptions = func() (operators []string) {
	for _, operator := range fastly.LogExplorerFilterOperators {
		operators = append(operators, string(operator))
	}
	return operators
}()

// Command exposes the Log Explorer API.
type Command struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	start string
	end   string

	// Optional.
	serviceName argparser.OptionalServiceNameID
	filters     []string
	limit       argparser.OptionalInt
	cursor      argparser.OptionalString
}

// NewLogExplorerCommand returns a usable Log Explorer command registered under the parent.
func NewLogExplorerCommand(parent argparser.Registerer, g *global.Data) *Command {
	c := Command{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(CommandName, "Retrieve sampled log records")

	// Required.
	c.CmdClause.Flag("start", "Inclusive start time in RFC3339 format").Required().StringVar(&c.start)
	c.CmdClause.Flag("end", "Exclusive end time in RFC3339 format").Required().StringVar(&c.end)

	// Optional.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("filter", "Filter in FIELD,OPERATOR,VALUE format (repeatable)").StringsVar(&c.filters)
	c.CmdClause.Flag("limit", "Maximum number of rows to return (up to 100)").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("cursor", "Pagination cursor from a previous response").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the Log Explorer API.
func (c *Command) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	filters, err := parseLogExplorerFilters(c.filters)
	if err != nil {
		return err
	}

	input := &fastly.GetLogRecordsInput{
		ServiceID: serviceID,
		Start:     c.start,
		End:       c.end,
		Filters:   filters,
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.cursor.WasSet {
		input.NextCursor = &c.cursor.Value
	}

	result, err := c.Globals.APIClient.GetLogRecords(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{"Service ID": serviceID})
		return err
	}

	if ok, err := c.WriteJSON(out, result); ok {
		return err
	}

	printLogRecords(out, result)
	return nil
}

func parseLogExplorerFilters(filters []string) ([]fastly.LogExplorerFilter, error) {
	result := make([]fastly.LogExplorerFilter, 0, len(filters))
	for _, filter := range filters {
		parts := strings.SplitN(filter, ",", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid --filter value %q: expected FIELD,OPERATOR,VALUE", filter)
		}

		field := strings.TrimSpace(parts[0])
		operator := strings.TrimSpace(parts[1])
		if field == "" || operator == "" {
			return nil, fmt.Errorf("invalid --filter value %q: field and operator must not be empty", filter)
		}
		if !containsLogExplorerOption(logExplorerFilterFieldOptions, field) {
			return nil, fmt.Errorf(
				"invalid --filter value %q: field must be one of [%s]",
				filter,
				strings.Join(logExplorerFilterFieldOptions, ", "),
			)
		}
		if !containsLogExplorerOption(logExplorerFilterOperatorOptions, operator) {
			return nil, fmt.Errorf(
				"invalid --filter value %q: operator must be one of [%s]",
				filter,
				strings.Join(logExplorerFilterOperatorOptions, ", "),
			)
		}

		result = append(result, fastly.LogExplorerFilter{
			Field:    fastly.LogExplorerFilterField(field),
			Operator: fastly.LogExplorerFilterOperator(operator),
			Value:    strings.TrimSpace(parts[2]),
		})
	}
	return result, nil
}

func containsLogExplorerOption(options []string, value string) bool {
	for _, option := range options {
		if option == value {
			return true
		}
	}
	return false
}

func printLogRecords(out io.Writer, result *fastly.LogRecordsResponse) {
	if result == nil || len(result.Data) == 0 {
		fmt.Fprintln(out, "No log records found.")
		printLogExplorerCursor(out, result)
		return
	}

	table := text.NewTable(out)
	table.AddHeader("TIMESTAMP", "METHOD", "HOST", "PATH", "STATUS", "POP", "CACHE HIT", "RESPONSE TIME")

	for _, record := range result.Data {
		if record == nil {
			continue
		}
		table.AddLine(
			logExplorerString(record.Timestamp),
			logExplorerString(record.RequestMethod),
			logExplorerString(record.RequestHost),
			logExplorerString(record.RequestPath),
			logExplorerInt(record.ResponseStatus),
			logExplorerString(record.FastlyPOP),
			logExplorerBool(record.IsCacheHit),
			logExplorerFloat(record.ResponseTime),
		)
	}
	table.Print()

	printLogExplorerCursor(out, result)
}

func printLogExplorerCursor(out io.Writer, result *fastly.LogRecordsResponse) {
	if result != nil && result.Meta != nil && result.Meta.NextCursor != nil && *result.Meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext cursor: %s\n", *result.Meta.NextCursor)
	}
}

func logExplorerString(v *string) string {
	if v == nil {
		return "-"
	}
	return *v
}

func logExplorerInt(v *int) any {
	if v == nil {
		return "-"
	}
	return *v
}

func logExplorerBool(v *bool) any {
	if v == nil {
		return "-"
	}
	return *v
}

func logExplorerFloat(v *float64) any {
	if v == nil {
		return "-"
	}
	return *v
}
