package item

import (
	"fmt"
	"io"
	"slices"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/common"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, globals *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a custom dashboard item").Alias("add")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("dashboard-id", "Alphanumeric string identifying a Dashboard").Required().StringVar(&c.dashboardID) // --dashboard-id
	c.CmdClause.Flag("item-id", "Alphanumeric string identifying a Dashboard Item").Required().StringVar(&c.itemID)      // --item-id

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag())                                                                                                                                                                                               // --json
	c.CmdClause.Flag("title", "A human-readable title for the dashboard item").Action(c.title.Set).StringVar(&c.title.Value)                                                                                                       // --title
	c.CmdClause.Flag("subtitle", "A human-readable subtitle for the dashboard item. Often a description of the visualization").Action(c.subtitle.Set).StringVar(&c.subtitle.Value)                                                 // --subtitle
	c.CmdClause.Flag("span", `The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on "desktop" screen sizes`).Action(c.span.Set).IntVar(&c.span.Value)                               // --span
	c.CmdClause.Flag("source-type", "The source of the data to display").Action(c.sourceType.Set).HintOptions(sourceTypes...).EnumVar(&c.sourceType.Value, sourceTypes...)                                                         // --source-type
	c.CmdClause.Flag("metrics", "The metrics to visualize. Valid options depend on the selected data source (set flag once per Metric)").Action(c.metrics.Set).StringsVar(&c.metrics.Value)                                        // --metrics
	c.CmdClause.Flag("visualization-type", `The type of visualization to display. Currently, only "chart" is supported`).Action(c.vizType.Set).HintOptions(visualizationTypes...).EnumVar(&c.vizType.Value, visualizationTypes...) // --visualization-type
	c.CmdClause.Flag("calculation-method", "The aggregation function to apply to the dataset").Action(c.calculationMethod.Set).HintOptions(calculationMethods...).EnumVar(&c.calculationMethod.Value, calculationMethods...)       // --calculation-method
	c.CmdClause.Flag("format", "The units to use to format the data").Action(c.format.Set).HintOptions(formats...).EnumVar(&c.format.Value, formats...)                                                                            // --format
	c.CmdClause.Flag("plot-type", "The type of chart to display").Action(c.plotType.Set).HintOptions(plotTypes...).EnumVar(&c.plotType.Value, plotTypes...)                                                                        // --plot-type

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// required
	dashboardID string
	itemID      string

	// optional
	title             argparser.OptionalString
	subtitle          argparser.OptionalString
	span              argparser.OptionalInt
	sourceType        argparser.OptionalString
	metrics           argparser.OptionalStringSlice
	plotType          argparser.OptionalString
	vizType           argparser.OptionalString
	calculationMethod argparser.OptionalString
	format            argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(&fastly.GetObservabilityCustomDashboardInput{ID: &c.dashboardID})
	if err != nil {
		return err
	}

	input, err := c.constructInput(d)
	if err != nil {
		return err
	}

	d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	// text.Success(out, `Added %d items to Custom Dashboard "%s" (id: %s)`, len(*input.Items), d.Name, d.ID)
	common.PrintDashboard(out, 0, d)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(d *fastly.ObservabilityCustomDashboard) (*fastly.UpdateObservabilityCustomDashboardInput, error) {
	var input fastly.UpdateObservabilityCustomDashboardInput

	input.ID = &d.ID
	input.Items = &d.Items
	idx := slices.IndexFunc(*input.Items, func(di fastly.DashboardItem) bool {
		return di.ID == c.itemID
	})
	if idx < 0 {
		return nil, fmt.Errorf("dashboard (%s) does not contain item with ID %s", d.ID, c.itemID)
	}
	item := &(*input.Items)[idx]

	if c.title.WasSet {
		(*item).Title = c.title.Value
	}
	if c.subtitle.WasSet {
		(*item).Subtitle = c.subtitle.Value
	}
	if c.span.WasSet {
		if span := c.span.Value; span <= 255 && span >= 0 {
			(*item).Span = uint8(span)
		} else {
			return nil, fmt.Errorf("invalid span value %d", span)
		}
	}
	if c.sourceType.WasSet {
		(*item).DataSource.Type = fastly.DashboardSourceType(c.sourceType.Value)
	}
	if c.metrics.WasSet {
		(*item).DataSource.Config.Metrics = c.metrics.Value
	}
	if c.vizType.WasSet {
		(*item).Visualization.Type = fastly.VisualizationType(c.vizType.Value)
	}
	if c.plotType.WasSet {
		(*item).Visualization.Config.PlotType = fastly.PlotType(c.plotType.Value)
	}
	if c.calculationMethod.WasSet {
		(*item).Visualization.Config.CalculationMethod = fastly.ToPointer(fastly.CalculationMethod(c.calculationMethod.Value))
	}
	if c.format.WasSet {
		(*item).Visualization.Config.Format = fastly.ToPointer(fastly.VisualizationFormat(c.format.Value))
	}

	return &input, nil
}
