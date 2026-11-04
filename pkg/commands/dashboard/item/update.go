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
	fsterr "github.com/fastly/cli/pkg/errors"
)

// AfterApply checks the values kong cannot check by itself.
func (c *UpdateCommand) AfterApply(ctx *kong.Context) error {
	if (c.SourceType.Value != "" || kongcompat.Given(ctx, "source-type")) && !slices.Contains([]string{"stats.domain", "stats.edge", "stats.origin"}, c.SourceType.Value) {
		return fmt.Errorf("enum value must be one of stats.domain,stats.edge,stats.origin, got '%s'", c.SourceType.Value)
	}
	if (c.VizType.Value != "" || kongcompat.Given(ctx, "visualization-type")) && !slices.Contains([]string{"chart"}, c.VizType.Value) {
		return fmt.Errorf("enum value must be one of chart, got '%s'", c.VizType.Value)
	}
	if (c.CalculationMethod.Value != "" || kongcompat.Given(ctx, "calculation-method")) && !slices.Contains([]string{"avg", "sum", "min", "max", "latest", "p95"}, c.CalculationMethod.Value) {
		return fmt.Errorf("enum value must be one of avg,sum,min,max,latest,p95, got '%s'", c.CalculationMethod.Value)
	}
	if (c.Format.Value != "" || kongcompat.Given(ctx, "format")) && !slices.Contains([]string{"number", "bytes", "percent", "requests", "responses", "seconds", "milliseconds", "ratio", "bitrate"}, c.Format.Value) {
		return fmt.Errorf("enum value must be one of number,bytes,percent,requests,responses,seconds,milliseconds,ratio,bitrate, got '%s'", c.Format.Value)
	}
	if (c.PlotType.Value != "" || kongcompat.Given(ctx, "plot-type")) && !slices.Contains([]string{"bar", "donut", "line", "single-metric"}, c.PlotType.Value) {
		return fmt.Errorf("enum value must be one of bar,donut,line,single-metric, got '%s'", c.PlotType.Value)
	}
	return nil
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required flags

	// required
	DashboardID string `required:"" help:"ID of the Dashboard containing the item"`
	ItemID      string `required:"" help:"ID of the Item to be updated"`

	// Optional flags
	argparser.JSONOutput

	// optional
	Title             argparser.OptionalString      `help:"A human-readable title for the dashboard item"`
	Subtitle          argparser.OptionalString      `help:"A human-readable subtitle for the dashboard item. Often a description of the visualization"`
	Span              argparser.OptionalInt         `help:"The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on \"desktop\" screen sizes"`
	SourceType        argparser.OptionalString      `hints:"stats.domain,stats.edge,stats.origin" help:"The source of the data to display"`
	Metrics           argparser.OptionalStringSlice `name:"metric" sep:"none" help:"The metrics to visualize. Valid options depend on the selected data source. Set flag multiple times to include multiple metrics"`
	VizType           argparser.OptionalString      `name:"visualization-type" hints:"chart" help:"The type of visualization to display. Currently, only \"chart\" is supported"`
	CalculationMethod argparser.OptionalString      `hints:"avg,sum,min,max,latest,p95" help:"The aggregation function to apply to the dataset"`
	Format            argparser.OptionalString      `hints:"number,bytes,percent,requests,responses,seconds,milliseconds,ratio,bitrate" help:"The units to use to format the data"`
	PlotType          argparser.OptionalString      `hints:"bar,donut,line,single-metric" help:"The type of chart to display"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(context.TODO(), &fastly.GetObservabilityCustomDashboardInput{ID: &c.DashboardID})
	if err != nil {
		return err
	}

	input, err := c.constructInput(d)
	if err != nil {
		return err
	}

	d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(d *fastly.ObservabilityCustomDashboard) (*fastly.UpdateObservabilityCustomDashboardInput, error) {
	var input fastly.UpdateObservabilityCustomDashboardInput

	input.ID = &d.ID
	input.Items = &d.Items
	idx := slices.IndexFunc(*input.Items, func(di fastly.DashboardItem) bool {
		return di.ID == c.ItemID
	})
	if idx < 0 {
		return nil, fmt.Errorf("dashboard (%s) does not contain item with ID %s", d.ID, c.ItemID)
	}
	item := &(*input.Items)[idx]

	if c.Title.WasSet {
		item.Title = c.Title.Value
	}
	if c.Subtitle.WasSet {
		item.Subtitle = c.Subtitle.Value
	}
	if c.Span.WasSet {
		if span := c.Span.Value; span <= 255 && span >= 0 {
			item.Span = uint8(span)
		} else {
			return nil, fmt.Errorf("invalid span value %d", span)
		}
	}
	if c.SourceType.WasSet {
		item.DataSource.Type = fastly.DashboardSourceType(c.SourceType.Value)
	}
	if c.Metrics.WasSet {
		item.DataSource.Config.Metrics = c.Metrics.Value
	}
	if c.VizType.WasSet {
		item.Visualization.Type = fastly.VisualizationType(c.VizType.Value)
	}
	if c.PlotType.WasSet {
		item.Visualization.Config.PlotType = fastly.PlotType(c.PlotType.Value)
	}
	if c.CalculationMethod.WasSet {
		item.Visualization.Config.CalculationMethod = new(fastly.CalculationMethod(c.CalculationMethod.Value))
	}
	if c.Format.WasSet {
		item.Visualization.Config.Format = new(fastly.VisualizationFormat(c.Format.Value))
	}

	return &input, nil
}
