package alerts_test

import (
	"strings"
	"testing"
	"time"

	root "github.com/fastly/cli/pkg/commands/alerts"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v9/fastly"
)

func TestAlertsCreate(t *testing.T) {
	var createFlags = flagList{
		Flags: []flag{
			{Flag: "--name", Value: "name"},
			{Flag: "--description", Value: "description"},
			{Flag: "--metric", Value: "status_5xx"},
			{Flag: "--source", Value: "stats"},
			{Flag: "--type", Value: "above_threshold"},
			{Flag: "--period", Value: "5m"},
			{Flag: "--threshold", Value: "10.0"},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "ok all required",
			Args: createFlags.String(),
			API:  mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
		{
			Name:      "no name",
			Args:      createFlags.Remove("--name").String(),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "no description",
			Args:      createFlags.Remove("--description").String(),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name:      "no metric",
			Args:      createFlags.Remove("--metric").String(),
			WantError: "error parsing arguments: required flag --metric not provided",
		},
		{
			Name:      "no source",
			Args:      createFlags.Remove("--source").String(),
			WantError: "error parsing arguments: required flag --source not provided",
		},
		{
			Name:      "no type",
			Args:      createFlags.Remove("--type").String(),
			WantError: "error parsing arguments: required flag --type not provided",
		},
		{
			Name:      "no period",
			Args:      createFlags.Remove("--period").String(),
			WantError: "error parsing arguments: required flag --period not provided",
		},
		{
			Name:      "no threshold",
			Args:      createFlags.Remove("--threshold").String(),
			WantError: "error parsing arguments: required flag --threshold not provided",
		},
		{
			Name: "ok optional json",
			Args: createFlags.Add(flag{Flag: "--json"}).String(),
			API:  mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
		{
			Name: "ok optional ignoreBelow",
			Args: createFlags.Add(flag{Flag: "--ignoreBelow", Value: "5.0"}).String(),
			API:  mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
		{
			Name: "ok optional service-id",
			Args: createFlags.Add(flag{Flag: "--service-id", Value: "ABC"}).String(),
			API:  mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
		{
			Name: "ok optional dimensions",
			Args: createFlags.
				Change(flag{Flag: "--source", Value: "origins"}).
				Add(flag{Flag: "--dimensions", Value: "fastly.com"}).
				Add(flag{Flag: "--dimensions", Value: "fastly2.com"}).String(),
			API: mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
		{
			Name: "ok optional integrations",
			Args: createFlags.
				Add(flag{Flag: "--integrations", Value: "ABC1"}).
				Add(flag{Flag: "--integrations", Value: "ABC2"}).String(),
			API: mock.API{CreateAlertDefinitionFn: CreateAlertDefinitionResponse},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestAlertsUpdate(t *testing.T) {
	var updateFlags = flagList{
		Flags: []flag{
			{Flag: "--id", Value: "ABC"},
			{Flag: "--name", Value: "name"},
			{Flag: "--description", Value: "description"},
			{Flag: "--metric", Value: "status_5xx"},
			{Flag: "--source", Value: "stats"},
			{Flag: "--type", Value: "above_threshold"},
			{Flag: "--period", Value: "5m"},
			{Flag: "--threshold", Value: "10.0"},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "ok all required",
			Args: updateFlags.String(),
			API:  mock.API{UpdateAlertDefinitionFn: UpdateAlertDefinitionResponse},
		},
		{
			Name:      "no id",
			Args:      updateFlags.Remove("--id").String(),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "no name",
			Args:      updateFlags.Remove("--name").String(),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "no description",
			Args:      updateFlags.Remove("--description").String(),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name:      "no metric",
			Args:      updateFlags.Remove("--metric").String(),
			WantError: "error parsing arguments: required flag --metric not provided",
		},
		{
			Name:      "no source",
			Args:      updateFlags.Remove("--source").String(),
			WantError: "error parsing arguments: required flag --source not provided",
		},
		{
			Name:      "no type",
			Args:      updateFlags.Remove("--type").String(),
			WantError: "error parsing arguments: required flag --type not provided",
		},
		{
			Name:      "no period",
			Args:      updateFlags.Remove("--period").String(),
			WantError: "error parsing arguments: required flag --period not provided",
		},
		{
			Name:      "no threshold",
			Args:      updateFlags.Remove("--threshold").String(),
			WantError: "error parsing arguments: required flag --threshold not provided",
		},
		{
			Name: "ok optional json",
			Args: updateFlags.Add(flag{Flag: "--json"}).String(),
			API:  mock.API{UpdateAlertDefinitionFn: UpdateAlertDefinitionResponse},
		},
		{
			Name: "ok optional ignoreBelow",
			Args: updateFlags.Add(flag{Flag: "--ignoreBelow", Value: "5.0"}).String(),
			API:  mock.API{UpdateAlertDefinitionFn: UpdateAlertDefinitionResponse},
		},
		{
			Name: "ok optional dimensions",
			Args: updateFlags.
				Change(flag{Flag: "--source", Value: "origins"}).
				Add(flag{Flag: "--dimensions", Value: "fastly.com"}).
				Add(flag{Flag: "--dimensions", Value: "fastly2.com"}).String(),
			API: mock.API{UpdateAlertDefinitionFn: UpdateAlertDefinitionResponse},
		},
		{
			Name: "ok optional integrations",
			Args: updateFlags.Add(flag{Flag: "--integrations", Value: "ABC1"}).Add(flag{Flag: "--integrations", Value: "ABC2"}).String(),
			API:  mock.API{UpdateAlertDefinitionFn: UpdateAlertDefinitionResponse},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func TestAlertsDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "no definition id",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "ok",
			Args: "--id ABC",
			API: mock.API{
				DeleteAlertDefinitionFn: func(i *fastly.DeleteAlertDefinitionInput) error {
					return nil
				},
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestAlertsDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "no definition id",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "ok",
			Args: "--id ABC",
			API: mock.API{
				GetAlertDefinitionFn: func(i *fastly.GetAlertDefinitionInput) (*fastly.AlertDefinition, error) {
					response := &mockDefinition
					return response, nil
				},
			},
			WantOutput: listAlertsOutput,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestAlertsList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:       "ok",
			API:        mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
			WantOutput: listAlertsEmptyOutput,
		},
		{
			Name: "ok verbose",
			Args: "-v",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok json",
			Args: "-j",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok cursor",
			Args: "--cursor ABC",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok limit",
			Args: "--limit 1",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok definition name",
			Args: "--name test",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok sort name",
			Args: "--sort name",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok sort updated_at",
			Args: "--sort updated_at",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok sort created_at asc",
			Args: "--sort created_at --order asc",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok sort created_at desc",
			Args: "--sort created_at --order desc",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "ok service id",
			Args: "--service-id ABC",
			API:  mock.API{ListAlertDefinitionsFn: ListAlertDefinitionsEmptyResponse},
		},
		{
			Name: "validate ListAlerts API success",
			API: mock.API{
				ListAlertDefinitionsFn: func(i *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error) {
					response := &fastly.AlertDefinitionsResponse{
						Data: []fastly.AlertDefinition{mockDefinition},
						Meta: fastly.AlertsMeta{
							Total:      1,
							Limit:      100,
							NextCursor: "",
							Sort:       "-name",
						},
					}
					return response, nil
				},
			},
			WantOutput: listAlertsOutput,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestAlertsHistoryList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:       "ok",
			API:        mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
			WantOutput: listAlertHistoryEmptyOutput,
		},
		{
			Name: "ok verbose",
			Args: "-v",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok json",
			Args: "--json",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok cursor",
			Args: "--cursor ABC",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok limit",
			Args: "--limit 1",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok status",
			Args: "--status active",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok sort start",
			Args: "--sort start",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok sort start asc",
			Args: "--sort start --order asc",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok sort start desc",
			Args: "--sort start --order desc",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok service id",
			Args: "--service-id ABC",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "ok definition id",
			Args: "--definition-id ABC",
			API:  mock.API{ListAlertHistoryFn: ListAlertHistoryEmptyResponse},
		},
		{
			Name: "validate ListAlerts API success",
			API: mock.API{
				ListAlertHistoryFn: func(i *fastly.ListAlertHistoryInput) (*fastly.AlertHistoryResponse, error) {
					response := &fastly.AlertHistoryResponse{
						Data: []fastly.AlertHistory{mockHistory},
						Meta: fastly.AlertsMeta{
							Total:      1,
							Limit:      100,
							NextCursor: "",
							Sort:       "-start",
						},
					}
					return response, nil
				},
			},
			WantOutput: listAlertsHistoryOutput,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "history"}, scenarios)
}

type flag struct {
	Flag  string
	Value string
}

func (t *flag) String() string {
	if t.Value == "" {
		return t.Flag
	}
	return strings.Join([]string{t.Flag, t.Value}, " ")
}

type flagList struct {
	Flags []flag
}

func (t *flagList) Add(flag flag) *flagList {
	newTuples := flagList{}
	newTuples.Flags = append(newTuples.Flags, t.Flags...)
	newTuples.Flags = append(newTuples.Flags, flag)
	return &newTuples
}

func (t *flagList) Change(flag flag) *flagList {
	newTuples := flagList{}
	for i := range t.Flags {
		if t.Flags[i].Flag == flag.Flag {
			newTuples.Flags = append(newTuples.Flags, flag)
		} else {
			newTuples.Flags = append(newTuples.Flags, t.Flags[i])

		}
	}
	return &newTuples
}

func (t *flagList) Remove(flag string) *flagList {
	newTuples := flagList{}
	for i := range t.Flags {
		if t.Flags[i].Flag != flag {
			newTuples.Flags = append(newTuples.Flags, t.Flags[i])
		}
	}
	return &newTuples
}
func (t *flagList) String() string {
	var strs []string
	for i := range t.Flags {
		strs = append(strs, t.Flags[i].String())
	}
	return strings.Join(strs, " ")
}

var mockTime = time.Date(2024, 05, 01, 12, 00, 11, 0, time.UTC)

var ListAlertDefinitionsEmptyResponse = func(i *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error) {
	response := &fastly.AlertDefinitionsResponse{
		Data: []fastly.AlertDefinition{},
		Meta: fastly.AlertsMeta{
			Total:      0,
			Limit:      100,
			NextCursor: "",
			Sort:       "-name",
		},
	}
	return response, nil
}

var ListAlertHistoryEmptyResponse = func(i *fastly.ListAlertHistoryInput) (*fastly.AlertHistoryResponse, error) {
	response := &fastly.AlertHistoryResponse{
		Data: []fastly.AlertHistory{},
		Meta: fastly.AlertsMeta{
			Total:      0,
			Limit:      100,
			NextCursor: "",
			Sort:       "-start",
		},
	}
	return response, nil
}

var mockDefinition = fastly.AlertDefinition{
	ID:             "ABC",
	Name:           "name",
	Description:    "description",
	Source:         "stats",
	Metric:         "status_5xx",
	ServiceID:      "SVC",
	Dimensions:     map[string][]string{},
	IntegrationIDs: []string{},
	EvaluationStrategy: map[string]any{
		"type":      "above_threshold",
		"period":    "5m",
		"threshold": 10.0,
	},
	UpdatedAt: mockTime,
	CreatedAt: mockTime,
}

var mockHistory = fastly.AlertHistory{
	ID:           "ABC",
	DefinitionID: mockDefinition.ID,
	Definition:   mockDefinition,
	Status:       "active",
	Start:        mockTime,
	End:          mockTime,
}

var CreateAlertDefinitionResponse = func(i *fastly.CreateAlertDefinitionInput) (*fastly.AlertDefinition, error) {
	response := &mockDefinition
	return response, nil
}

var UpdateAlertDefinitionResponse = func(i *fastly.UpdateAlertDefinitionInput) (*fastly.AlertDefinition, error) {
	response := &mockDefinition
	return response, nil
}

var listAlertsEmptyOutput = `DEFINITION ID  SERVICE ID  NAME  SOURCE  METRIC  TYPE  THRESHOLD  PERIOD`

var listAlertsOutput = `DEFINITION ID  SERVICE ID  NAME  SOURCE  METRIC      TYPE             THRESHOLD  PERIOD
ABC            SVC         name  stats   status_5xx  above_threshold  10         5m
`

var listAlertHistoryEmptyOutput = `HISTORY ID  DEFINITION ID  STATUS  START  END`

var listAlertsHistoryOutput = `HISTORY ID  DEFINITION ID  STATUS  START                          END
ABC         ABC            active  2024-05-01 12:00:11 +0000 UTC  2024-05-01 12:00:11 +0000 UTC
`
