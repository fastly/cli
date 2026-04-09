package stats_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAggregate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			name:       "success table",
			args:       args("stats aggregate"),
			api:        mock.API{GetAggregateJSONFn: getAggregateJSONOK},
			wantOutput: "From:",
		},
		{
			name:       "success json",
			args:       args("stats aggregate --format=json"),
			api:        mock.API{GetAggregateJSONFn: getAggregateJSONOK},
			wantOutput: `"start_time":0`,
		},
		{
			name:       "success json alias",
			args:       args("stats aggregate --json"),
			api:        mock.API{GetAggregateJSONFn: getAggregateJSONOK},
			wantOutput: `"start_time":0`,
		},
		{
			name:      "verbose json combo",
			args:      args("stats aggregate --json --verbose"),
			api:       mock.API{GetAggregateJSONFn: getAggregateJSONOK},
			wantError: "invalid flag combination",
		},
		{
			name:      "non-success status",
			args:      args("stats aggregate"),
			api:       mock.API{GetAggregateJSONFn: getAggregateJSONNonSuccess},
			wantError: "non-success response",
		},
		{
			name:      "api error",
			args:      args("stats aggregate"),
			api:       mock.API{GetAggregateJSONFn: getAggregateJSONError},
			wantError: errTest.Error(),
		},
	}
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(tc.args, &stdout)
				opts.APIClientFactory = mock.APIClient(tc.api)
				return opts, nil
			}
			err := app.Run(tc.args, nil)
			testutil.AssertErrorContains(t, err, tc.wantError)
			testutil.AssertStringContains(t, stdout.String(), tc.wantOutput)
		})
	}
}

func getAggregateJSONOK(_ context.Context, _ *fastly.GetAggregateInput, o any) error {
	msg := []byte(`{
  "status": "success",
  "meta": {"to": "Thu May 16 20:08:35 UTC 2013", "from": "Wed May 15 20:08:35 UTC 2013", "by": "day", "region": "all"},
  "msg": null,
  "data": [{"start_time": 0}]
}`)
	return json.Unmarshal(msg, o)
}

func getAggregateJSONNonSuccess(_ context.Context, _ *fastly.GetAggregateInput, o any) error {
	msg := []byte(`{"status": "error", "msg": "bad request", "meta": {}, "data": []}`)
	return json.Unmarshal(msg, o)
}

func getAggregateJSONError(_ context.Context, _ *fastly.GetAggregateInput, _ any) error {
	return errTest
}
