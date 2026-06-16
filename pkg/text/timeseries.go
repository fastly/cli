package text

import (
	"fmt"
	"io"
	"sort"

	"github.com/fastly/go-fastly/v15/fastly/ngwaf/v1/timeseries"
	wsts "github.com/fastly/go-fastly/v15/fastly/ngwaf/v1/workspaces/timeseries"
)

// PrintTimeseries displays timeseries data points in a table.
func PrintTimeseries(out io.Writer, t *timeseries.Timeseries) {
	if len(t.Data) == 0 {
		fmt.Fprintf(out, "Total: 0\n")
		return
	}

	// Collect sorted metric keys from the first data point. This allows us to dynamically adjust the data column label.
	var metricKeys []string
	if len(t.Data[0].Values) > 0 {
		for k := range t.Data[0].Values[0] {
			metricKeys = append(metricKeys, k)
		}
		sort.Strings(metricKeys)
	}

	headers := append([]string{"Time", "Workspace"}, metricKeys...)
	tbl := NewTable(out)
	tbl.AddHeader(toAny(headers)...)

	for _, dp := range t.Data {
		row := []any{dp.Dimensions.Time, dp.Dimensions.Workspace}
		if len(dp.Values) > 0 {
			for _, k := range metricKeys {
				row = append(row, dp.Values[0][k])
			}
		}
		tbl.AddLine(row...)
	}
	tbl.Print()

	fmt.Fprintf(out, "\nTotal: %d\n", t.Meta.Total)
}

func toAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// PrintWorkspaceTimeseries displays workspace timeseries data points in a table.
func PrintWorkspaceTimeseries(out io.Writer, t *wsts.TimeSeries) {
	if t.Meta.Total == 0 {
		fmt.Fprintf(out, "Total: 0\n")
		return
	}

	var metricKeys []string
	seen := map[string]bool{}
	for _, dp := range t.Data {
		for k := range dp {
			if k != "timestamp" && !seen[k] {
				seen[k] = true
				metricKeys = append(metricKeys, k)
			}
		}
	}
	sort.Strings(metricKeys)

	headers := append([]string{"Timestamp"}, metricKeys...)
	tbl := NewTable(out)
	tbl.AddHeader(toAny(headers)...)

	for _, dp := range t.Data {
		row := []any{dp["timestamp"]}
		for _, k := range metricKeys {
			row = append(row, dp[k])
		}
		tbl.AddLine(row...)
	}
	tbl.Print()

	fmt.Fprintf(out, "\nTotal: %d\n", t.Meta.Total)
}
