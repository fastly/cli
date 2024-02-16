package stats_test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
)

func TestHistorical(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       args("stats historical --service-id=123"),
			api:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			wantOutput: historicalOK,
		},
		{
			args:      args("stats historical --service-id=123"),
			api:       mock.API{GetStatsJSONFn: getStatsJSONError},
			wantError: errTest.Error(),
		},
		{
			args:       args("stats historical --service-id=123 --format=json"),
			api:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			wantOutput: historicalJSONOK,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var historicalOK = `From: Wed May 15 20:08:35 UTC 2013
To: Thu May 16 20:08:35 UTC 2013
By: day
Region: all
---
Service ID:                                    123
Start Time:          1970-01-01 00:00:00 +0000 UTC
--------------------------------------------------
Hit Rate:                                    0.00%
Avg Hit Time:                               0.00µs
Avg Miss Time:                              0.00µs

Request BW:                                      0
  Headers:                                       0
  Body:                                          0

Response BW:                                     0
  Headers:                                       0
  Body:                                          0

Requests:                                        0
  Hit:                                           0
  Miss:                                          0
  Pass:                                          0
  Synth:                                         0
  Error:                                         0
  Uncacheable:                                   0
`

var historicalJSONOK = `{"start_time":0}
`

func getStatsJSONOK(_ *fastly.GetStatsInput, o any) error {
	msg := []byte(`
{
  "status": "success",
  "meta": {
    "to": "Thu May 16 20:08:35 UTC 2013",
    "from": "Wed May 15 20:08:35 UTC 2013",
    "by": "day",
    "region": "all"
  },
  "msg": null,
  "data": [{"start_time": 0}]
}`)

	return json.Unmarshal(msg, o)
}

func getStatsJSONError(_ *fastly.GetStatsInput, _ any) error {
	return errTest
}
