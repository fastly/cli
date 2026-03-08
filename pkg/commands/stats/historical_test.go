package stats_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/stats"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHistorical(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:       "success",
			Args:       "--service-id=123",
			API:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			WantOutput: historicalOK,
		},
		{
			Name:      "api failure",
			Args:      "--service-id=123",
			API:       mock.API{GetStatsJSONFn: getStatsJSONError},
			WantError: errTest.Error(),
		},
		{
			Name:       "success with json format",
			Args:       "--service-id=123 --format=json",
			API:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			WantOutput: historicalJSONOK,
		},
		{
			Name:       "success with field filter",
			Args:       "--service-id=123 --field=bandwidth",
			API:        mock.API{GetStatsJSONFn: getStatsJSONFieldOK},
			WantOutput: historicalOK,
		},
		{
			Name:       "success with field filter and json format",
			Args:       "--service-id=123 --field=bandwidth --format=json",
			API:        mock.API{GetStatsJSONFn: getStatsJSONFieldOK},
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

func unmarshalStatsJSON(o any) error {
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

func getStatsJSONOK(_ context.Context, _ *fastly.GetStatsInput, o any) error {
	return unmarshalStatsJSON(o)
}

func getStatsJSONFieldOK(_ context.Context, i *fastly.GetStatsInput, o any) error {
	if i.Field == nil || *i.Field != "bandwidth" {
		return errTest
	}
	return unmarshalStatsJSON(o)
}

func getStatsJSONError(_ context.Context, _ *fastly.GetStatsInput, _ any) error {
	return errTest
}
