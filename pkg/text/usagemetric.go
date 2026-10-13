package text

import (
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/usagemetrics"
)

// PrintUsageMetricsTbl displays ARC usage metrics in a table format.
func PrintUsageMetricsTbl(out io.Writer, metricsToPrint []usagemetrics.UsageMetric) {
	tbl := NewTable(out)
	tbl.AddHeader("Date", "Usage Type", "Quantity", "Virtual Key", "Provider", "Model")

	if metricsToPrint == nil {
		tbl.Print()
		return
	}

	for _, m := range metricsToPrint {
		tbl.AddLine(m.Date, m.UsageType, strconv.Itoa(m.Quantity), m.VirtualKeyName, m.Provider, m.Model)
	}
	tbl.Print()
}
