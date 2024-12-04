package dashboard

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, ds []*fastly.ObservabilityCustomDashboard) {
	t := text.NewTable(out)
	t.AddHeader("DASHBOARD ID", "NAME", "DESCRIPTION", "# ITEMS")
	for _, d := range ds {
		t.AddLine(
			d.ID,
			d.Name,
			d.Description,
			len(d.Items),
		)
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, ds []*fastly.ObservabilityCustomDashboard) {
	for _, d := range ds {
		printDashboard(out, 0, d)
		fmt.Fprintf(out, "\n")
	}
}

func printDashboard(out io.Writer, indent uint, dashboard *fastly.ObservabilityCustomDashboard) {
	indentStep := uint(4)
	level := indent
	text.Indent(out, level, "Name: %s", dashboard.Name)
	text.Indent(out, level, "Description: %s", dashboard.Description)
	text.Indent(out, level, "Items:")

	level += indentStep
	for i, di := range dashboard.Items {
		text.Indent(out, level, "[%d]:", i)
		level += indentStep
		printItem(out, level, &di)
		level -= indentStep
	}
	level -= indentStep

	text.Indent(out, level, "Meta:")
	level += indentStep
	text.Indent(out, level, "Created at: %s", dashboard.CreatedAt)
	text.Indent(out, level, "Updated at: %s", dashboard.UpdatedAt)
	text.Indent(out, level, "Created by: %s", dashboard.CreatedBy)
	text.Indent(out, level, "Updated by: %s", dashboard.UpdatedBy)
}

func printItem(out io.Writer, indent uint, item *fastly.DashboardItem) {
	indentStep := uint(4)
	level := indent
	if item != nil {
		text.Indent(out, level, "ID: %s", item.ID)
		text.Indent(out, level, "Title: %s", item.Title)
		text.Indent(out, level, "Subtitle: %s", item.Subtitle)
		text.Indent(out, level, "Span: %d", item.Span)

		text.Indent(out, level, "Data Source:")
		level += indentStep
		text.Indent(out, level, "Type: %s", item.DataSource.Type)
		text.Indent(out, level, "Metrics: %s", strings.Join(item.DataSource.Config.Metrics, ", "))
		level -= indentStep

		text.Indent(out, level, "Visualization:")
		level += indentStep
		text.Indent(out, level, "Type: %s", item.Visualization.Type)
		text.Indent(out, level, "Plot Type: %s", item.Visualization.Config.PlotType)
		if item.Visualization.Config.CalculationMethod != nil {
			text.Indent(out, level, "Calculation Method: %s", *item.Visualization.Config.CalculationMethod)
		}
		if item.Visualization.Config.Format != nil {
			text.Indent(out, level, "Format: %s", *item.Visualization.Config.Format)
		}
	}
}
