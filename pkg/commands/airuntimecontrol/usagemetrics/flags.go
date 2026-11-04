package usagemetrics

import (
	"time"

	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/usagemetrics"

	"github.com/fastly/cli/pkg/argparser"
)

// filterFlags holds the query filters shared by the list and export
// operations.
type filterFlags struct {
	Key      argparser.OptionalString `help:"Filter by a specific virtual key ID"`
	Provider argparser.OptionalString `help:"Filter by provider connection ID"`
	Model    argparser.OptionalString `help:"Filter by AI model identifier"`
	From     time.Time                `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"Start of the time range (RFC 3339 format)"`
	To       time.Time                `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"End of the time range (RFC 3339 format). Defaults to now"`
	Sort     argparser.OptionalString `default:"date" placeholder:"\"date\"" help:"Sort field. Prefix with '-' for descending order (e.g. \"-date\")"`
}

// input builds a usagemetrics.ListInput from the parsed flag values.
func (f *filterFlags) input() *usagemetrics.ListInput {
	input := &usagemetrics.ListInput{}
	if f.Key.WasSet {
		input.Key = &f.Key.Value
	}
	if f.Provider.WasSet {
		input.Provider = &f.Provider.Value
	}
	if f.Model.WasSet {
		input.Model = &f.Model.Value
	}
	if !f.From.IsZero() {
		input.From = &f.From
	}
	if !f.To.IsZero() {
		input.To = &f.To
	}
	if f.Sort.WasSet {
		input.Sort = &f.Sort.Value
	}
	return input
}

// exportInput builds a usagemetrics.ExportInput from the parsed flag values.
func (f *filterFlags) exportInput() *usagemetrics.ExportInput {
	i := f.input()
	return &usagemetrics.ExportInput{
		Key:      i.Key,
		Provider: i.Provider,
		Model:    i.Model,
		From:     i.From,
		To:       i.To,
		Sort:     i.Sort,
	}
}
