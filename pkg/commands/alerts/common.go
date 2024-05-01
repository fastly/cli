package alerts

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly"
)

func printDefinition(out io.Writer, indent uint, definition *fastly.AlertDefinition) {
	if definition != nil {
		text.Indent(out, indent, "Definition ID: %s", fastly.ToValue(&definition.ID))
		text.Indent(out, indent, "Service ID: %s", fastly.ToValue(&definition.ServiceID))
		text.Indent(out, indent, "Name: %s", fastly.ToValue(&definition.Name))
		text.Indent(out, indent, "Source: %s", fastly.ToValue(&definition.Source))

		dimensions, ok := definition.Dimensions[fastly.ToValue(&definition.Source)]
		if ok && len(dimensions) > 0 {
			text.Indent(out, indent, "Dimensions:")
			for i := range dimensions {
				text.Indent(out, indent+4, "  %s", dimensions[i])
			}
		}

		text.Indent(out, indent, "Metric: %s", fastly.ToValue(&definition.Metric))

		text.Indent(out, indent, "Evaluation Strategy:")
		eType := definition.EvaluationStrategy["type"].(string)
		text.Indent(out, indent+4, "  Type: %s", fastly.ToValue(&eType))

		period := definition.EvaluationStrategy["period"].(string)
		text.Indent(out, indent+4, "  Period: %s", fastly.ToValue(&period))

		threshold := definition.EvaluationStrategy["threshold"].(float64)
		text.Indent(out, indent+4, "  Threshold: %v", fastly.ToValue(&threshold))

		if ignoreBelow, ok := definition.EvaluationStrategy["ignore_below"].(float64); ok {
			text.Indent(out, indent+4, "  IgnoreBelow: %v", fastly.ToValue(&ignoreBelow))
		}

		integrations := definition.IntegrationIDs
		if ok && len(integrations) > 0 {
			text.Indent(out, indent, "Integrations:")
			for i := range integrations {
				text.Indent(out, indent, "  %s", integrations[i])
			}
		}

		text.Indent(out, indent, "Created at: %s", definition.CreatedAt)
		text.Indent(out, indent, "Updated at: %s", definition.UpdatedAt)
	}
}

func printCursor(out io.Writer, meta *fastly.AlertsMeta) {
	if meta != nil && meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext Cursor: %s\n", fastly.ToValue(&meta.NextCursor))
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, meta *fastly.AlertsMeta, as []*fastly.AlertDefinition) {
	t := text.NewTable(out)
	t.AddHeader("DEFINITION ID", "SERVICE ID", "NAME", "SOURCE", "METRIC", "TYPE", "THRESHOLD", "PERIOD")
	for _, a := range as {
		eType := a.EvaluationStrategy["type"].(string)
		period := a.EvaluationStrategy["period"].(string)
		threshold := a.EvaluationStrategy["threshold"].(float64)
		t.AddLine(
			fastly.ToValue(&a.ID),
			fastly.ToValue(&a.ServiceID),
			fastly.ToValue(&a.Name),
			fastly.ToValue(&a.Source),
			fastly.ToValue(&a.Metric),
			fastly.ToValue(&eType),
			fastly.ToValue(&threshold),
			fastly.ToValue(&period),
		)
	}
	t.Print()
	printCursor(out, meta)
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, meta *fastly.AlertsMeta, as []*fastly.AlertDefinition) {
	for _, a := range as {
		printDefinition(out, 0, a)
		fmt.Fprintf(out, "\n")
	}
	printCursor(out, meta)
}

func printHistory(out io.Writer, history *fastly.AlertHistory) {
	if history != nil {
		start := history.Start.UTC().String()
		end := history.End.UTC().String()
		fmt.Fprintf(out, "History ID: %s\n", fastly.ToValue(&history.ID))
		fmt.Fprintf(out, "Definition:\n")
		printDefinition(out, 4, &history.Definition)
		fmt.Fprintf(out, "Status: %s\n", fastly.ToValue(&history.Status))
		fmt.Fprintf(out, "Start: %s\n", fastly.ToValue(&start))
		fmt.Fprintf(out, "End: %s\n", fastly.ToValue(&end))
		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func printHistorySummary(out io.Writer, meta *fastly.AlertsMeta, as []*fastly.AlertHistory) {
	t := text.NewTable(out)
	t.AddHeader("HISTORY ID", "DEFINITION ID", "STATUS", "START", "END")
	for _, a := range as {
		start := a.Start.UTC().String()
		end := a.End.UTC().String()
		t.AddLine(
			fastly.ToValue(&a.ID),
			fastly.ToValue(&a.DefinitionID),
			fastly.ToValue(&a.Status),
			fastly.ToValue(&start),
			fastly.ToValue(&end),
		)
	}
	t.Print()
	printCursor(out, meta)
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printHistoryVerbose(out io.Writer, meta *fastly.AlertsMeta, history []*fastly.AlertHistory) {
	for _, h := range history {
		printHistory(out, h)
	}
	printCursor(out, meta)
}
