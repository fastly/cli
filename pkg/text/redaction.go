package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/redactions"
)

// PrintRedaction displays a single redaction.
func PrintRedaction(out io.Writer, redactionToPrint *redactions.Redaction) {
	fmt.Fprintf(out, "Field: %s\n", redactionToPrint.Field)
	fmt.Fprintf(out, "ID: %s\n", redactionToPrint.RedactionID)
	fmt.Fprintf(out, "Type: %s\n", redactionToPrint.Type)
	fmt.Fprintf(out, "Created At: %s\n", redactionToPrint.CreatedAt)
}

// PrintRedactionTbl prints a table of redactions.
func PrintRedactionTbl(out io.Writer, redactionsToPrint []redactions.Redaction) {
	tbl := NewTable(out)
	tbl.AddHeader("Field", "ID", "Type", "Created At")

	if redactionsToPrint == nil {
		tbl.Print()
		return
	}

	for _, rd := range redactionsToPrint {
		tbl.AddLine(rd.Field, rd.RedactionID, rd.Type, rd.CreatedAt)
	}
	tbl.Print()
}
