package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
)

// PrintRule displays an NGWAF rule.
func PrintRule(out io.Writer, ruleToPrint *rules.Rule) {
	fmt.Fprintf(out, "ID: %s\n", ruleToPrint.RuleID)
	fmt.Fprintf(out, "Action: %s\n", ruleToPrint.Actions[0].Type)
	fmt.Fprintf(out, "Description: %s\n", ruleToPrint.Description)
	fmt.Fprintf(out, "Enabled: %v\n", ruleToPrint.Enabled)
	fmt.Fprintf(out, "Type: %s\n", ruleToPrint.Type)
	fmt.Fprintf(out, "Scope: %s\n", ruleToPrint.Scope.Type)
	fmt.Fprintf(out, "Updated (UTC): %s\n", ruleToPrint.UpdatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Created (UTC): %s\n", ruleToPrint.CreatedAt.UTC().Format(time.Format))
}

// PrintRuleTbl displays rules in a table format.
func PrintRuleTbl(out io.Writer, rulesToPrint []rules.Rule) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Action", "Description", "Enabled", "Type", "Scope", "Updated At", "Created At")

	if rulesToPrint == nil {
		tbl.Print()
		return
	}

	for _, ruleToPrint := range rulesToPrint {
		tbl.AddLine(
			ruleToPrint.RuleID,
			ruleToPrint.Actions[0].Type,
			ruleToPrint.Description,
			ruleToPrint.Enabled,
			ruleToPrint.Type,
			ruleToPrint.Scope.Type,
			ruleToPrint.UpdatedAt,
			ruleToPrint.CreatedAt,
		)
	}
	tbl.Print()
}
