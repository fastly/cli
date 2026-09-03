package usagemetrics

import (
	"time"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/usagemetrics"

	"github.com/fastly/cli/pkg/argparser"
)

// filterFlags holds the query filters shared by the list and export
// operations.
type filterFlags struct {
	key      argparser.OptionalString
	provider argparser.OptionalString
	model    argparser.OptionalString
	from     time.Time
	to       time.Time
	sort     argparser.OptionalString
}

// register defines the shared filter flags against the given command clause.
func (f *filterFlags) register(c argparser.Base) {
	c.CmdClause.Flag("key", "Filter by a specific virtual key ID").Action(f.key.Set).StringVar(&f.key.Value)
	c.CmdClause.Flag("provider", "Filter by provider name").Action(f.provider.Set).StringVar(&f.provider.Value)
	c.CmdClause.Flag("model", "Filter by model name").Action(f.model.Set).StringVar(&f.model.Value)
	c.CmdClause.Flag("from", "Start of the time range (RFC 3339 format)").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &f.from)
	c.CmdClause.Flag("to", "End of the time range (RFC 3339 format). Defaults to now").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &f.to)
	c.CmdClause.Flag("sort", "Sort field. Prefix with '-' for descending order (e.g. \"-date\")").Default("date").Action(f.sort.Set).StringVar(&f.sort.Value)
}

// input builds a usagemetrics.ListInput from the parsed flag values.
func (f *filterFlags) input() *usagemetrics.ListInput {
	input := &usagemetrics.ListInput{}
	if f.key.WasSet {
		input.Key = &f.key.Value
	}
	if f.provider.WasSet {
		input.Provider = &f.provider.Value
	}
	if f.model.WasSet {
		input.Model = &f.model.Value
	}
	if !f.from.IsZero() {
		input.From = &f.from
	}
	if !f.to.IsZero() {
		input.To = &f.to
	}
	if f.sort.WasSet {
		input.Sort = &f.sort.Value
	}
	return input
}
