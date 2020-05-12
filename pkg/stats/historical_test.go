package stats_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/fastly"
)

func TestHistorical(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"stats", "historical"},
			api:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			wantOutput: historicalOK,
		},
		{
			args:      []string{"stats", "historical"},
			api:       mock.API{GetStatsJSONFn: getStatsJSONError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"stats", "historical", "--format=json"},
			api:        mock.API{GetStatsJSONFn: getStatsJSONOK},
			wantOutput: historicalJSONOK,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var historicalOK = `From: Wed May 15 20:08:35 UTC 2013
To: Thu May 16 20:08:35 UTC 2013
By: day
Region: all
---
Service ID:                              serviceID
Start Time:          1969-12-31 16:00:00 -0800 PST
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

func getStatsJSONOK(i *fastly.GetStatsInput, o interface{}) error {
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
  "data": {
    "serviceID": [{"start_time": 0}]
  }
}`)

	return json.Unmarshal(msg, o)
}

func getStatsJSONError(i *fastly.GetStatsInput, o interface{}) error {
	return errTest
}
