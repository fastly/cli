package item

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/printer"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply checks the values kong cannot check by itself.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.CalculationMethod.Value != "" || kongcompat.Given(ctx, "calculation-method")) && !slices.Contains([]string{"avg", "sum", "min", "max", "latest", "p95"}, c.CalculationMethod.Value) {
		return fmt.Errorf("enum value must be one of avg,sum,min,max,latest,p95, got '%s'", c.CalculationMethod.Value)
	}
	if (c.Format.Value != "" || kongcompat.Given(ctx, "format")) && !slices.Contains([]string{"number", "bytes", "percent", "requests", "responses", "seconds", "milliseconds", "ratio", "bitrate"}, c.Format.Value) {
		return fmt.Errorf("enum value must be one of number,bytes,percent,requests,responses,seconds,milliseconds,ratio,bitrate, got '%s'", c.Format.Value)
	}
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required flags

	// required
	DashboardID string   `required:"" help:"ID of the Dashboard to contain the item"`
	Title       string   `required:"" help:"A human-readable title for the dashboard item"`
	Subtitle    string   `required:"" help:"A human-readable subtitle for the dashboard item. Often a description of the visualization"`
	SourceType  string   `required:"" enum:"stats.domain,stats.edge,stats.origin" hints:"stats.domain,stats.edge,stats.origin" help:"The source of the data to display"`
	Metrics     []string `name:"metric" required:"" sep:"none" help:"The metrics to visualize. Valid options depend on the selected data source. Set flag multiple times to include multiple metrics"`
	PlotType    string   `required:"" enum:"bar,donut,line,single-metric" hints:"bar,donut,line,single-metric" help:"The type of chart to display"`

	// Optional flags
	argparser.JSONOutput

	// optional
	VizType           string                   `name:"visualization-type" default:"chart" enum:"chart" hints:"chart" help:"The type of visualization to display. Currently, only \"chart\" is supported"`
	CalculationMethod argparser.OptionalString `hints:"avg,sum,min,max,latest,p95" help:"The aggregation function to apply to the dataset"`
	Format            argparser.OptionalString `hints:"number,bytes,percent,requests,responses,seconds,milliseconds,ratio,bitrate" help:"The units to use to format the data"`
	Span              uint8                    `default:"4" help:"The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on \"desktop\" screen sizes"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(context.TODO(), &fastly.GetObservabilityCustomDashboardInput{ID: &c.DashboardID})
	if err != nil {
		return err
	}

	input := c.constructInput(d)
	d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	text.Success(out, `Added item to Custom Dashboard "%s" (id: %s)`, d.Name, d.ID)
	// Summary isn't useful for a single dashboard, so print verbose by default
	printer.PrintDashboard(out, 0, d)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(d *fastly.ObservabilityCustomDashboard) *fastly.UpdateObservabilityCustomDashboardInput {
	input := fastly.UpdateObservabilityCustomDashboardInput{
		ID:          &d.ID,
		Name:        &d.Name,
		Description: &d.Description,
		Items:       &d.Items,
	}
	item := fastly.DashboardItem{
		Title:    c.Title,
		Subtitle: c.Subtitle,
		Span:     c.Span,
		DataSource: fastly.DashboardDataSource{
			Type: fastly.DashboardSourceType(c.SourceType),
			Config: fastly.DashboardSourceConfig{
				Metrics: c.Metrics,
			},
		},
		Visualization: fastly.DashboardVisualization{
			Type: fastly.VisualizationType(c.VizType),
			Config: fastly.VisualizationConfig{
				PlotType: fastly.PlotType(c.PlotType),
			},
		},
	}
	if c.CalculationMethod.WasSet {
		item.Visualization.Config.CalculationMethod = new(fastly.CalculationMethod(c.CalculationMethod.Value))
	}
	if c.Format.WasSet {
		item.Visualization.Config.Format = new(fastly.VisualizationFormat(c.Format.Value))
	}

	*input.Items = append(*input.Items, item)

	return &input
}
