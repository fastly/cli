package beacon_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/internal/beacon"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNotify(t *testing.T) {
	args := testutil.SplitArgs("compute deploy")
	out := bytes.NewBuffer(nil)
	g := testutil.MockGlobalData(args, out)
	m := &mockHTTPClient{
		resp: &http.Response{
			StatusCode: http.StatusNoContent,
			Status:     http.StatusText(http.StatusNoContent),
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}
	g.HTTPClient = m

	err := beacon.Notify(g, "service-id", beacon.Event{
		Name:   "test-event",
		Status: beacon.StatusSuccess,
	})

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "/cli/service-id/notify", m.req.URL.Path)
	testutil.AssertEqual(t, "fastly-notification-relay.edgecompute.app", m.req.URL.Host)

	rawData, err := io.ReadAll(m.req.Body)
	testutil.AssertNoError(t, err)
	defer m.req.Body.Close()

	var data map[string]any
	err = json.Unmarshal(rawData, &data)
	testutil.AssertNoError(t, err)

	name, ok := data["event"].(string)
	testutil.AssertBool(t, true, ok)
	testutil.AssertEqual(t, "test-event", name)

	result, ok := data["status"].(string)
	testutil.AssertBool(t, true, ok)
	testutil.AssertEqual(t, "success", result)
}

type mockHTTPClient struct {
	req  *http.Request
	resp *http.Response
	err  error
}

func (m *mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	m.req = r
	return m.resp, m.err
}
