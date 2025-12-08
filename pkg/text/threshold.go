package text

import (
	"fmt"
	"io"
	"time"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/thresholds"
)

// PrintThreshold displays a single redaction.
func PrintThreshold(out io.Writer, thresholdToPrint *thresholds.Threshold) {
	fmt.Fprintf(out, "Signal: %s\n", thresholdToPrint.Signal)
	fmt.Fprintf(out, "Name: %s\n", thresholdToPrint.Name)
	fmt.Fprintf(out, "Action: %s\n", thresholdToPrint.Action)
	fmt.Fprintf(out, "Do Not Notify: %t\n", thresholdToPrint.DontNotify)
	fmt.Fprintf(out, "Duration: %d\n", thresholdToPrint.Duration)
	fmt.Fprintf(out, "Enabled: %t\n", thresholdToPrint.Enabled)
	fmt.Fprintf(out, "Interval: %d\n", thresholdToPrint.Interval)
	fmt.Fprintf(out, "Limit: %d\n", thresholdToPrint.Limit)

}

// PrintThresholdTbl prints a table of thresholds.
func PrintThresholdTbl(out io.Writer, thresholdsToPrint []thresholds.Threshold) {
	tbl := NewTable(out)
	tbl.AddHeader("Signal", "Name", "ID", "Action", "Enabled", "Do Not Notify", "Limit", "Interval", "Duration", "Created At")

	if thresholdsToPrint == nil {
		tbl.Print()
		return
	}

	for _, ts := range thresholdsToPrint {
		tbl.AddLine(
			ts.Signal,
			ts.Name,
			ts.ThresholdID,
			ts.Action,
			ts.Enabled,
			ts.DontNotify,
			ts.Limit,
			ts.Interval,
			ts.Duration,
			ts.CreatedAt.UTC().Format(time.RFC3339),
		)
	}
	tbl.Print()
}
