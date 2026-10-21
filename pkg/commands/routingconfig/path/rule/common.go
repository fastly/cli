package rule

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, data []rules.Data) {
	t := text.NewTable(out)
	t.AddHeader("RULE ID", "ACTION TYPE", "ACTION VALUE", "IS DEFAULT", "CONDITIONS")
	for _, d := range data {
		t.AddLine(d.RuleID, d.Action.Type, d.Action.Value, d.IsDefault, len(d.Conditions))
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, data []rules.Data) {
	for _, d := range data {
		fmt.Fprintf(out, "Rule ID: %s\n", d.RuleID)
		fmt.Fprintf(out, "Action Type: %s\n", d.Action.Type)
		fmt.Fprintf(out, "Action Value: %s\n", d.Action.Value)
		fmt.Fprintf(out, "Is Default: %t\n", d.IsDefault)
		for _, cond := range d.Conditions {
			key := ""
			if cond.Key != nil {
				key = *cond.Key
			}
			fmt.Fprintf(out, "Condition: type=%s, key=%s, operator=%s, value=%s\n", cond.Type, key, cond.Operator, cond.Value)
		}
		fmt.Fprintf(out, "Created at: %s\n", d.CreatedAt)
		fmt.Fprintf(out, "Updated at: %s\n", d.UpdatedAt)
		fmt.Fprintf(out, "\n")
	}
}
