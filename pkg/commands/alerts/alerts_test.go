package alerts_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v9/fastly"
)

func TestAlertsList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "ok",
			Args: args("alerts list"),
		},
		{
			Name: "validate ListAlerts API success",
			API: mock.API{
				ListAlertDefinitionsFn: func(i *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error) {
					response := &fastly.AlertDefinitionsResponse{
						Data: []fastly.AlertDefinition{
							{},
						},
						Meta: fastly.AlertsMeta{
							Total:      1,
							Limit:      10,
							NextCursor: "",
							Sort:       "-name",
						},
					}
					return response, nil
				},
			},
			Args:       args("alerts list"),
			WantOutput: listAlertsEntriesOutput,
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

var listAlertsEntriesOutput = `SERVICE ID  ID   IP         SUBNET  NEGATED
123         456  127.0.0.1  0       false
123         789  127.0.0.2  0       true
`
