package stats_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	root "github.com/fastly/cli/pkg/commands/stats"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHistorical(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:       "--service-id=123",
			API:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			WantOutput: historicalOK,
		},
		{
			Args:      "--service-id=123",
			API:       mock.API{GetStatsJSONFn: getStatsJSONError},
			WantError: errTest.Error(),
		},
		{
			Args:       "--service-id=123 --format=json",
			API:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			WantOutput: historicalJSONOK,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "historical"}, scenarios)
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

func getStatsJSONOK(_ context.Context, _ *fastly.GetStatsInput, o any) error {
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

func getStatsJSONError(_ context.Context, _ *fastly.GetStatsInput, _ any) error {
	return errTest
}
