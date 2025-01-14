package dashboard_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/dashboard"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	userID = "test-user"
)

func TestCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate CreateObservabilityCustomDashboard API error",
			API: mock.API{
				CreateObservabilityCustomDashboardFn: func(i *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name Testing",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate missing --name flag",
			API: mock.API{
				CreateObservabilityCustomDashboardFn: func(i *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return nil, testutil.Err
				},
			},
			Args:      "",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate optional --description flag",
			API: mock.API{
				CreateObservabilityCustomDashboardFn: func(i *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return &fastly.ObservabilityCustomDashboard{
						ID:   "beepboop",
						Name: i.Name,
					}, nil
				},
			},
			Args:       "--name Testing",
			WantOutput: `Created Custom Dashboard "Testing" (id: beepboop)`,
		},
		{
			Name: "validate CreateObservabilityCustomDashboard API success",
			API: mock.API{
				CreateObservabilityCustomDashboardFn: func(i *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return &fastly.ObservabilityCustomDashboard{
						ID:          "beepboop",
						Name:        i.Name,
						Description: *i.Description,
					}, nil
				},
			},
			Args:       "--name Testing --description foo",
			WantOutput: `Created Custom Dashboard "Testing" (id: beepboop)`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate DeleteObservabilityCustomDashboard API error",
			API: mock.API{
				DeleteObservabilityCustomDashboardFn: func(i *fastly.DeleteObservabilityCustomDashboardInput) error {
					return testutil.Err
				},
			},
			Args:      "--id beepboop",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteObservabilityCustomDashboard API success",
			API: mock.API{
				DeleteObservabilityCustomDashboardFn: func(i *fastly.DeleteObservabilityCustomDashboardInput) error {
					return nil
				},
			},
			Args:       "--id beepboop",
			WantOutput: "Deleted Custom Dashboard beepboop",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate GetObservabilityCustomDashboard API error",
			API: mock.API{
				GetObservabilityCustomDashboardFn: func(i *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id beepboop",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetObservabilityCustomDashboard API success",
			API: mock.API{
				GetObservabilityCustomDashboardFn: getObservabilityCustomDashboard,
			},
			Args:       "--id beepboop",
			WantOutput: "Name: Testing\nDescription: This is a test dashboard\nItems:\nMeta:\n    Created at: 2021-06-15 23:00:00 +0000 UTC\n    Updated at: 2021-06-15 23:00:00 +0000 UTC\n    Created by: test-user\n    Updated by: test-user\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate ListObservabilityCustomDashboards API error",
			API: mock.API{
				ListObservabilityCustomDashboardsFn: func(i *fastly.ListObservabilityCustomDashboardsInput) (*fastly.ListDashboardsResponse, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListObservabilityCustomDashboards API success",
			API: mock.API{
				ListObservabilityCustomDashboardsFn: listObservabilityCustomDashboards,
			},
			WantOutput: "DASHBOARD ID  NAME       DESCRIPTION  # ITEMS\nbeepboop      Testing 1  This is #1   0\nbleepblorp    Testing 2  This is #2   0\n",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListObservabilityCustomDashboardsFn: listObservabilityCustomDashboards,
			},
			Args:       "--verbose",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nName: Testing 1\nDescription: This is #1\nItems:\nMeta:\n    Created at: 2021-06-15 23:00:00 +0000 UTC\n    Updated at: 2021-06-15 23:00:00 +0000 UTC\n    Created by: test-user\n    Updated by: test-user\n\nName: Testing 2\nDescription: This is #2\nItems:\nMeta:\n    Created at: 2021-06-15 23:00:00 +0000 UTC\n    Updated at: 2021-06-15 23:00:00 +0000 UTC\n    Created by: test-user\n    Updated by: test-user\n\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate UpdateObservabilityCustomDashboard API error",
			API: mock.API{
				UpdateObservabilityCustomDashboardFn: func(i *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id beepboop",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateObservabilityCustomDashboard API success",
			API: mock.API{
				UpdateObservabilityCustomDashboardFn: func(i *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
					return &fastly.ObservabilityCustomDashboard{
						ID:          *i.ID,
						Name:        *i.Name,
						Description: *i.Description,
					}, nil
				},
			},
			Args:       "--id beepboop --name Foo --description Bleepblorp",
			WantOutput: "SUCCESS: Updated Custom Dashboard \"Foo\" (id: beepboop)\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func getObservabilityCustomDashboard(i *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	t := testutil.Date

	return &fastly.ObservabilityCustomDashboard{
		CreatedAt:   t,
		CreatedBy:   userID,
		Description: "This is a test dashboard",
		ID:          *i.ID,
		Items:       []fastly.DashboardItem{},
		Name:        "Testing",
		UpdatedAt:   t,
		UpdatedBy:   userID,
	}, nil
}

func listObservabilityCustomDashboards(i *fastly.ListObservabilityCustomDashboardsInput) (*fastly.ListDashboardsResponse, error) {
	t := testutil.Date
	vs := &fastly.ListDashboardsResponse{
		Data: []fastly.ObservabilityCustomDashboard{{
			CreatedAt:   t,
			CreatedBy:   userID,
			Description: "This is #1",
			ID:          "beepboop",
			Items:       []fastly.DashboardItem{},
			Name:        "Testing 1",
			UpdatedAt:   t,
			UpdatedBy:   userID,
		}, {
			CreatedAt:   t,
			CreatedBy:   userID,
			Description: "This is #2",
			ID:          "bleepblorp",
			Items:       []fastly.DashboardItem{},
			Name:        "Testing 2",
			UpdatedAt:   t,
			UpdatedBy:   userID,
		}},
		Meta: fastly.DashboardMeta{},
	}

	return vs, nil
}
