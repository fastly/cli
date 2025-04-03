package item

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/common"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, globals *global.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a custom dashboard item").Alias("add")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("dashboard-id", "ID of the Dashboard to contain the item").Required().StringVar(&c.dashboardID)
	c.CmdClause.Flag("title", "A human-readable title for the dashboard item").Required().StringVar(&c.title)
	c.CmdClause.Flag("subtitle", "A human-readable subtitle for the dashboard item. Often a description of the visualization").Required().StringVar(&c.subtitle)
	c.CmdClause.Flag("source-type", "The source of the data to display").Required().HintOptions(sourceTypes...).EnumVar(&c.sourceType, sourceTypes...)
	c.CmdClause.Flag("metric", "The metrics to visualize. Valid options depend on the selected data source. Set flag multiple times to include multiple metrics").Required().StringsVar(&c.metrics)
	c.CmdClause.Flag("plot-type", "The type of chart to display").Required().HintOptions(plotTypes...).EnumVar(&c.plotType, plotTypes...)

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("visualization-type", `The type of visualization to display. Currently, only "chart" is supported`).Default("chart").HintOptions(visualizationTypes...).EnumVar(&c.vizType, visualizationTypes...)
	c.CmdClause.Flag("calculation-method", "The aggregation function to apply to the dataset").Action(c.calculationMethod.Set).HintOptions(calculationMethods...).EnumVar(&c.calculationMethod.Value, calculationMethods...) // --calculation-method
	c.CmdClause.Flag("format", "The units to use to format the data").Action(c.format.Set).HintOptions(formats...).EnumVar(&c.format.Value, formats...)                                                                      // --format
	c.CmdClause.Flag("span", `The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on "desktop" screen sizes`).Default("4").Uint8Var(&c.span)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// required
	dashboardID string
	title       string
	subtitle    string
	sourceType  string
	metrics     []string
	plotType    string

	// optional
	vizType           string
	calculationMethod argparser.OptionalString
	format            argparser.OptionalString
	span              uint8
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(&fastly.GetObservabilityCustomDashboardInput{ID: &c.dashboardID})
	if err != nil {
		return err
	}

	input := c.constructInput(d)
	d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	text.Success(out, `Added item to Custom Dashboard "%s" (id: %s)`, d.Name, d.ID)
	// Summary isn't useful for a single dashboard, so print verbose by default
	common.PrintDashboard(out, 0, d)
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
		Title:    c.title,
		Subtitle: c.subtitle,
		Span:     c.span,
		DataSource: fastly.DashboardDataSource{
			Type: fastly.DashboardSourceType(c.sourceType),
			Config: fastly.DashboardSourceConfig{
				Metrics: c.metrics,
			},
		},
		Visualization: fastly.DashboardVisualization{
			Type: fastly.VisualizationType(c.vizType),
			Config: fastly.VisualizationConfig{
				PlotType: fastly.PlotType(c.plotType),
			},
		},
	}
	if c.calculationMethod.WasSet {
		item.Visualization.Config.CalculationMethod = fastly.ToPointer(fastly.CalculationMethod(c.calculationMethod.Value))
	}
	if c.format.WasSet {
		item.Visualization.Config.Format = fastly.ToPointer(fastly.VisualizationFormat(c.format.Value))
	}

	*input.Items = append(*input.Items, item)

	return &input
}
