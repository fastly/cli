package text

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v17/fastly"

	fsttime "github.com/fastly/cli/pkg/time"
)

func fmtIntegrationTime(t *time.Time) string {
	if t == nil {
		return "n/a"
	}
	return t.UTC().Format(fsttime.Format)
}

// PrintIntegrationsTbl displays integrations in a table format.
func PrintIntegrationsTbl(out io.Writer, integrations []fastly.Integration) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "ID", "Type", "Status", "Created (UTC)", "Updated (UTC)")
	for _, i := range integrations {
		tbl.AddLine(
			fastly.ToValue(i.Name),
			fastly.ToValue(i.ID),
			fastly.ToValue(i.Type),
			fastly.ToValue(i.Status),
			fmtIntegrationTime(i.CreatedAt),
			fmtIntegrationTime(i.UpdatedAt),
		)
	}
	tbl.Print()
}

// PrintIntegration displays detailed information about a single integration.
func PrintIntegration(out io.Writer, i *fastly.Integration) {
	PrintLines(out, Lines{
		"Name":          fastly.ToValue(i.Name),
		"ID":            fastly.ToValue(i.ID),
		"Type":          fastly.ToValue(i.Type),
		"Status":        fastly.ToValue(i.Status),
		"Description":   fastly.ToValue(i.Description),
		"Created (UTC)": fmtIntegrationTime(i.CreatedAt),
		"Updated (UTC)": fmtIntegrationTime(i.UpdatedAt),
	})

	if len(i.Config) == 0 {
		return
	}

	keys := make([]string, 0, len(i.Config))
	for k := range i.Config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprint(out, "Config:\n")
	for _, k := range keys {
		fmt.Fprintf(out, "\t%s: %s\n", k, i.Config[k])
	}
}

// PrintIntegrationTypesTbl displays the supported integration types and their
// configuration fields in a table format.
func PrintIntegrationTypesTbl(out io.Writer, types []fastly.IntegrationType) {
	tbl := NewTable(out)
	tbl.AddHeader("Type", "Display Name", "Custom Fields")
	for _, t := range types {
		fields := make([]string, 0, len(t.CustomFields))
		for _, f := range t.CustomFields {
			name := fastly.ToValue(f.Name)
			if format := fastly.ToValue(f.Format); format != "" {
				name = fmt.Sprintf("%s (%s)", name, format)
			}
			fields = append(fields, name)
		}
		tbl.AddLine(fastly.ToValue(t.Type), fastly.ToValue(t.DisplayName), strings.Join(fields, ", "))
	}
	tbl.Print()
}
