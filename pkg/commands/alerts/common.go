package alerts

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly"
)

const (
	defaultEvaluationType   = "above_threshold"
	defaultEvaluationPeriod = "5m"
)

// evaluationType is a list of supported evaluation types.
var evaluationType = []string{"above_threshold", "all_above_threshold", "below_threshold", "percent_absolute", "percent_decrease", "percent_increase"}

// evaluationPeriod is a list of supported evaluation periods.
var evaluationPeriod = []string{"2m", "3m", "5m", "15m", "30m"}

func printDefinition(out io.Writer, indent uint, definition *fastly.AlertDefinition) {
	if definition != nil {
		text.Indent(out, indent, "Definition ID: %s", definition.ID)
		text.Indent(out, indent, "Service ID: %s", definition.ServiceID)
		text.Indent(out, indent, "Name: %s", definition.Name)
		text.Indent(out, indent, "Source: %s", definition.Source)

		dimensions, ok := definition.Dimensions[definition.Source]
		if ok && len(dimensions) > 0 {
			text.Indent(out, indent, "Dimensions:")
			for i := range dimensions {
				text.Indent(out, indent+4, "  %s", dimensions[i])
			}
		}

		text.Indent(out, indent, "Metric: %s", definition.Metric)

		text.Indent(out, indent, "Evaluation Strategy:")
		eType := definition.EvaluationStrategy["type"].(string)
		text.Indent(out, indent+4, "  Type: %s", eType)

		period := definition.EvaluationStrategy["period"].(string)
		text.Indent(out, indent+4, "  Period: %s", period)

		threshold := definition.EvaluationStrategy["threshold"].(float64)
		text.Indent(out, indent+4, "  Threshold: %v", threshold)

		if ignoreBelow, ok := definition.EvaluationStrategy["ignore_below"].(float64); ok {
			text.Indent(out, indent+4, "  IgnoreBelow: %v", ignoreBelow)
		}

		integrations := definition.IntegrationIDs
		if len(integrations) > 0 {
			text.Indent(out, indent, "Integrations:")
			for i := range integrations {
				text.Indent(out, indent, "  %s", integrations[i])
			}
		}

		text.Indent(out, indent, "Created at: %s", definition.CreatedAt)
		text.Indent(out, indent, "Updated at: %s", definition.UpdatedAt)
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, as []*fastly.AlertDefinition) {
	t := text.NewTable(out)
	t.AddHeader("DEFINITION ID", "SERVICE ID", "NAME", "SOURCE", "METRIC", "TYPE", "THRESHOLD", "PERIOD")
	for _, a := range as {
		eType := a.EvaluationStrategy["type"].(string)
		period := a.EvaluationStrategy["period"].(string)
		threshold := a.EvaluationStrategy["threshold"].(float64)
		t.AddLine(
			a.ID,
			a.ServiceID,
			a.Name,
			a.Source,
			a.Metric,
			eType,
			threshold,
			period,
		)
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, as []*fastly.AlertDefinition) {
	for _, a := range as {
		printDefinition(out, 0, a)
		fmt.Fprintf(out, "\n")
	}
}

func printHistory(out io.Writer, history *fastly.AlertHistory) {
	if history != nil {
		start := history.Start.UTC().String()
		end := history.End.UTC().String()
		fmt.Fprintf(out, "History ID: %s\n", history.ID)
		fmt.Fprintf(out, "Definition:\n")
		printDefinition(out, 4, &history.Definition)
		fmt.Fprintf(out, "Status: %s\n", history.Status)
		fmt.Fprintf(out, "Start: %s\n", start)
		fmt.Fprintf(out, "End: %s\n", end)
		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func printHistorySummary(out io.Writer, as []*fastly.AlertHistory) {
	t := text.NewTable(out)
	t.AddHeader("HISTORY ID", "DEFINITION ID", "STATUS", "START", "END")
	for _, a := range as {
		start := a.Start.UTC().String()
		end := a.End.UTC().String()
		t.AddLine(
			a.ID,
			a.DefinitionID,
			a.Status,
			start,
			end,
		)
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printHistoryVerbose(out io.Writer, history []*fastly.AlertHistory) {
	for _, h := range history {
		printHistory(out, h)
	}
}
