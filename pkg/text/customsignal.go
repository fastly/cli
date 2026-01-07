package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"
)

// PrintCustomSignal displays an NGWAF custom signal.
func PrintCustomSignal(out io.Writer, customSignalToPrint *signals.Signal) {
	fmt.Fprintf(out, "ID: %s\n", customSignalToPrint.SignalID)
	fmt.Fprintf(out, "Name: %s\n", customSignalToPrint.Name)
	fmt.Fprintf(out, "Description: %s\n", customSignalToPrint.Description)
	fmt.Fprintf(out, "Scope: %s\n", customSignalToPrint.Scope.Type)
	fmt.Fprintf(out, "Updated (UTC): %s\n", customSignalToPrint.UpdatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Created (UTC): %s\n", customSignalToPrint.CreatedAt.UTC().Format(time.Format))
}

// PrintCustomSignalTbl displays custom signals in a table format.
func PrintCustomSignalTbl(out io.Writer, customSignalsToPrint []signals.Signal) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Description", "Scope", "Updated At", "Created At")

	if customSignalsToPrint == nil {
		tbl.Print()
		return
	}

	for _, customSignalToPrint := range customSignalsToPrint {
		tbl.AddLine(
			customSignalToPrint.SignalID,
			customSignalToPrint.Name,
			customSignalToPrint.Description,
			customSignalToPrint.Scope.Type,
			customSignalToPrint.UpdatedAt,
			customSignalToPrint.CreatedAt,
		)
	}
	tbl.Print()
}
