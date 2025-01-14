package item_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/dashboard"
	sub "github.com/fastly/cli/pkg/commands/dashboard/item"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

var (
	testDate             = testutil.Date
	userID               = "test-user"
	dashboardID          = "beepboop"
	itemID               = "bleepblorp"
	dashboardName        = "Foo"
	dashboardDescription = "Testing..."
	title                = "Title"
	subtitle             = "Subtitle"
	sourceType           = "stats.edge"
	metrics              = "requests"
	plotType             = "line"
	vizType              = "chart"
	calculationMethod    = "latest"
	format               = "requests"
	span                 = 8
	defaultItem          = fastly.DashboardItem{
		DataSource: fastly.DashboardDataSource{
			Config: fastly.DashboardSourceConfig{
				Metrics: []string{metrics},
			},
			Type: fastly.DashboardSourceType(sourceType),
		},
		ID:       itemID,
		Span:     uint8(span),
		Subtitle: subtitle,
		Title:    title,
		Visualization: fastly.DashboardVisualization{
			Config: fastly.VisualizationConfig{
				CalculationMethod: fastly.ToPointer(fastly.CalculationMethod(calculationMethod)),
				Format:            fastly.ToPointer(fastly.VisualizationFormat(format)),
				PlotType:          fastly.PlotType(plotType),
			},
			Type: fastly.VisualizationType(vizType),
		},
	}
	defaultDashboard = func() fastly.ObservabilityCustomDashboard {
		return fastly.ObservabilityCustomDashboard{
			CreatedAt:   testDate,
			CreatedBy:   userID,
			Description: dashboardDescription,
			ID:          dashboardID,
			Items:       []fastly.DashboardItem{defaultItem},
			Name:        dashboardName,
			UpdatedAt:   testDate,
			UpdatedBy:   userID,
		}
	}
)

func TestCreate(t *testing.T) {
	allRequiredFlags := fmt.Sprintf("--dashboard-id %s --title %s --subtitle %s --source-type %s --metric %s --plot-type %s", dashboardID, title, subtitle, sourceType, metrics, plotType)
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --dashboard-id flag",
			Args:      fmt.Sprintf("--title %s --subtitle %s --source-type %s --metric %s --plot-type %s", title, subtitle, sourceType, metrics, plotType),
			WantError: "error parsing arguments: required flag --dashboard-id not provided",
		},
		{
			Name:      "validate missing --title flag",
			Args:      fmt.Sprintf("--dashboard-id %s --subtitle %s --source-type %s --metric %s --plot-type %s", dashboardID, subtitle, sourceType, metrics, plotType),
			WantError: "error parsing arguments: required flag --title not provided",
		},
		{
			Name:      "validate missing --subtitle flag",
			Args:      fmt.Sprintf("--dashboard-id %s --title %s --source-type %s --metric %s --plot-type %s", dashboardID, title, sourceType, metrics, plotType),
			WantError: "error parsing arguments: required flag --subtitle not provided",
		},
		{
			Name:      "validate missing --source-type flag",
			Args:      fmt.Sprintf("--dashboard-id %s --title %s --subtitle %s --metric %s --plot-type %s", dashboardID, title, subtitle, metrics, plotType),
			WantError: "error parsing arguments: required flag --source-type not provided",
		},
		{
			Name:      "validate missing --metric flag",
			Args:      fmt.Sprintf("--dashboard-id %s --title %s --subtitle %s --source-type %s --plot-type %s", dashboardID, title, subtitle, sourceType, plotType),
			WantError: "error parsing arguments: required flag --metric not provided",
		},
		{
			Name:      "validate missing --plot-type flag",
			Args:      fmt.Sprintf("--dashboard-id %s --title %s --subtitle %s --source-type %s --metric %s", dashboardID, title, subtitle, sourceType, metrics),
			WantError: "error parsing arguments: required flag --plot-type not provided",
		},
		{
			Name: "validate multiple --metric flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       allRequiredFlags + " --metric responses",
			WantOutput: "Metrics: requests, responses",
		},
		{
			Name: "validate all required flags",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       allRequiredFlags,
			WantOutput: `Added item to Custom Dashboard "Foo"`,
		},
		{
			Name: "validate all optional flags",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --visualization-type %s --calculation-method %s --format %s --span %d", allRequiredFlags, vizType, calculationMethod, format, span),
			WantOutput: `Added item to Custom Dashboard "Foo"`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDelete(t *testing.T) {
	allRequiredFlags := fmt.Sprintf("--dashboard-id %s --item-id %s", dashboardID, itemID)
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --dashboard-id flag",
			Args:      fmt.Sprintf("--item-id %s", itemID),
			WantError: "error parsing arguments: required flag --dashboard-id not provided",
		},
		{
			Name:      "validate missing --item-id flag",
			Args:      fmt.Sprintf("--dashboard-id %s", dashboardID),
			WantError: "error parsing arguments: required flag --item-id not provided",
		},
		{
			Name: "validate all required flags",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardEmpty,
			},
			Args:       allRequiredFlags,
			WantOutput: `Removed 1 dashboard item(s) from Custom Dashboard "Foo"`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestDescribe(t *testing.T) {
	allRequiredFlags := fmt.Sprintf("--dashboard-id %s --item-id %s", dashboardID, itemID)
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --dashboard-id flag",
			Args:      fmt.Sprintf("--item-id %s", itemID),
			WantError: "error parsing arguments: required flag --dashboard-id not provided",
		},
		{
			Name:      "validate missing --item-id flag",
			Args:      fmt.Sprintf("--dashboard-id %s", dashboardID),
			WantError: "error parsing arguments: required flag --item-id not provided",
		},
		{
			Name: "validate all required flags",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardEmpty,
			},
			Args:       allRequiredFlags,
			WantOutput: "ID: bleepblorp\nTitle: Title\nSubtitle: Subtitle\nSpan: 8\nData Source:\n    Type: stats.edge\n    Metrics: requests\nVisualization:\n    Type: chart\n    Plot Type: line\n    Calculation Method: latest\n    Format: requests\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestUpdate(t *testing.T) {
	allRequiredFlags := fmt.Sprintf("--dashboard-id %s --item-id %s --json", dashboardID, itemID)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --dashboard-id flag",
			Args:      fmt.Sprintf("--item-id %s", itemID),
			WantError: "error parsing arguments: required flag --dashboard-id not provided",
		},
		{
			Name:      "validate missing --item-id flag",
			Args:      fmt.Sprintf("--dashboard-id %s", dashboardID),
			WantError: "error parsing arguments: required flag --item-id not provided",
		},
		{
			Name: "validate all required flags",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args: allRequiredFlags,
			WantOutputs: []string{
				`"name":`, "Foo",
				`"description":`, "Testing...",
				`"items":`,
				`"id":`, "bleepblorp",
				`"title":`, "Title",
				`"subtitle":`, "Subtitle",
				`"span":`, "8",
				`"data_source":`,
				`"type":`, "stats.edge",
				`"metrics":`, "requests",
				`"visualization":`,
				`"type":`, "chart",
				`"plot_type":`, "line",
				`"calculation_method":`, "latest",
				`"format":`, "requests",
				`"created_at":`, "2021-06-15T23:00:00Z",
				`"updated_at":`, "2021-06-15T23:00:00Z",
				`"created_by":`, "test-user",
				`"updated_by":`, "test-user",
			},
		},
		{
			Name: "validate optional --title flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --title %s", allRequiredFlags, "NewTitle"),
			WantOutput: `"title": "NewTitle"`,
		},
		{
			Name: "validate optional --subtitle flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --subtitle %s", allRequiredFlags, "NewSubtitle"),
			WantOutput: `"subtitle": "NewSubtitle"`,
		},
		{
			Name: "validate optional --span flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --span %d", allRequiredFlags, 12),
			WantOutput: `"span": 12`,
		},
		{
			Name: "validate optional --source-type flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --source-type %s", allRequiredFlags, "stats.domain"),
			WantOutput: `"type": "stats.domain"`,
		},
		{
			Name: "validate optional --metric flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:        fmt.Sprintf("%s --metric %s", allRequiredFlags, "status_4xx"),
			WantOutputs: []string{"metrics", "status_4xx"},
		},
		{
			Name: "validate multiple --metric flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args: fmt.Sprintf("%s --metric %s --metric %s --metric %s", allRequiredFlags, "status_2xx", "status_4xx", "status_5xx"),
			WantOutputs: []string{
				"metrics",
				"status_2xx",
				"status_4xx",
				"status_5xx",
			},
		},
		{
			Name: "validate optional --calculation-method flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --calculation-method %s", allRequiredFlags, "avg"),
			WantOutput: `"calculation_method": "avg"`,
		},
		{
			Name: "validate optional --format flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --format %s", allRequiredFlags, "ratio"),
			WantOutput: `"format": "ratio"`,
		},
		{
			Name: "validate optional --plot-type flag",
			API: mock.API{
				GetObservabilityCustomDashboardFn:    getDashboardOK,
				UpdateObservabilityCustomDashboardFn: updateDashboardOK,
			},
			Args:       fmt.Sprintf("%s --plot-type %s", allRequiredFlags, "single-metric"),
			WantOutput: `"plot_type": "single-metric"`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func getDashboardOK(i *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	d := defaultDashboard()
	return &d, nil
}

func updateDashboardOK(i *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	d := defaultDashboard()
	d.Items = *i.Items
	return &d, nil
}

func updateDashboardEmpty(i *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	d := defaultDashboard()
	d.Items = []fastly.DashboardItem{}
	return &d, nil
}
