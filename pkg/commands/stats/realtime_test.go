package stats_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRealtime(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		name      string
		args      []string
		api       mock.API
		wantError string
	}{
		{
			name:      "verbose json combo",
			args:      args("stats realtime --service-id 123 --json --verbose"),
			api:       mock.API{},
			wantError: "invalid flag combination",
		},
		{
			name:      "verbose format json combo",
			args:      args("stats realtime --service-id 123 --format=json --verbose"),
			api:       mock.API{},
			wantError: "invalid flag combination",
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
		})
	}
}
