package category_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/category"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
)

const categoryPath = "api.example.com/bot-management/v1/workspaces/ws123/policy/categories/search-engines"

// categoryTransport verifies each request and rejects unexpected API calls.
type categoryTransport struct {
	t     *testing.T
	steps []categoryRequest
	next  int
}

type categoryRequest struct {
	method   string
	path     string
	body     string
	cursor   string
	response string
	status   int
}

func (c *categoryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.t.Helper()
	if c.next >= len(c.steps) {
		c.t.Errorf("unexpected request: %s %s", req.Method, req.URL)
		return nil, fmt.Errorf("unexpected request")
	}
	step := c.steps[c.next]
	c.next++
	var wantBody *string
	if step.body != "" {
		wantBody = &step.body
	}
	testutil.AssertRequest(c.t, req, testutil.ExpectedRequest{
		Method: step.method, Path: step.path, WantJSON: wantBody,
	})
	if got := req.URL.Query().Get("cursor"); got != step.cursor {
		c.t.Errorf("cursor = %q, want %q", got, step.cursor)
	}
	status := step.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(step.response)),
	}, nil
}

func categoryClient(t *testing.T, steps []categoryRequest) *http.Client {
	t.Helper()
	transport := &categoryTransport{t: t, steps: steps}
	t.Cleanup(func() {
		if transport.next != len(steps) {
			t.Errorf("made %d API calls, want %d", transport.next, len(steps))
		}
	})
	return &http.Client{Transport: transport}
}

func TestCategoryClearOverrides(t *testing.T) {
	updated := category
	updated.Action.Type = "challenge"
	for _, tc := range []struct {
		name  string
		flags string
		clear bool
		json  bool
	}{
		{name: "preserves overrides by default"},
		{name: "clears overrides when requested", flags: " --clear-overrides", clear: true},
		{name: "clears overrides with JSON output", flags: " --clear-overrides --json", clear: true, json: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []categoryRequest{{
				method: http.MethodPatch, path: categoryPath,
				body: `{"action":{"type":"challenge"}}`, response: string(testutil.GenJSON(updated)),
			}}
			if tc.clear {
				steps = append(steps, categoryRequest{
					method: http.MethodPatch, path: categoryPath + "/bots",
					body: `{"action":{"type":"inherit"}}`, response: `{"data":[]}`,
				})
			}
			scenario := testutil.CLIScenario{
				Name:       tc.name,
				Args:       "--workspace-id ws123 --category-id search-engines --action challenge" + tc.flags,
				Client:     categoryClient(t, steps),
				WantOutput: "Updated bot category",
			}
			if tc.json {
				scenario.WantOutput = fstfmt.EncodeJSON(updated)
			}
			testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, "update"}, scenario)
		})
	}
}

func TestCategoryClearOverridesErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		steps       []categoryRequest
		remediation string
	}{
		{
			name: "failed category update does not clear overrides",
			steps: []categoryRequest{{
				method: http.MethodPatch, path: categoryPath,
				body: `{"action":{"type":"block"}}`, status: http.StatusNotFound,
				response: `{"title":"not found","status":404}`,
			}},
		},
		{
			name: "failed clearing reports partial success",
			steps: []categoryRequest{
				{
					method: http.MethodPatch, path: categoryPath,
					body: `{"action":{"type":"block"}}`, response: string(testutil.GenJSON(category)),
				},
				{
					method: http.MethodPatch, path: categoryPath + "/bots",
					body: `{"action":{"type":"inherit"}}`, status: http.StatusNotFound,
					response: `{"title":"not found","status":404}`,
				},
			},
			remediation: "The category action was updated, but the API call for clearing bot overrides failed. Rerun the command with --clear-overrides to try again.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, "update"}, testutil.CLIScenario{
				Name: tc.name, Args: "--workspace-id ws123 --category-id search-engines --action block --clear-overrides",
				Client: categoryClient(t, tc.steps), WantError: "404 - Not Found",
				WantRemediation: tc.remediation, DontWantOutput: "Updated bot category",
			})
		})
	}
}

func TestCategoryOverrideCount(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		pages   []string
		want    string
	}{
		{
			name: "no overrides", command: "describe",
			pages: []string{`{"data":[{"action":{"type":"inherit"}}]}`},
			want:  "Action: allow\nOverrides: 0",
		},
		{
			name: "one override even when matching category action", command: "get",
			pages: []string{`{"data":[{"action":{"type":"allow"}},{"action":{"type":"inherit"}}]}`},
			want:  "Action: allow\nOverrides: 1",
		},
		{
			name: "counts overrides across all pages", command: "describe",
			pages: []string{
				`{"data":[{"action":{"type":"allow"}},{"action":{"type":"inherit"}}],"meta":{"next_cursor":"next"}}`,
				`{"data":[{"action":{"type":"block"}},{"action":{"type":"challenge"}}]}`,
			},
			want: "Action: allow\nOverrides: 3",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []categoryRequest{{method: http.MethodGet, path: categoryPath, response: string(testutil.GenJSON(category))}}
			for i, page := range tc.pages {
				cursor := ""
				if i > 0 {
					cursor = "next"
				}
				steps = append(steps, categoryRequest{method: http.MethodGet, path: categoryPath + "/bots", response: page, cursor: cursor})
			}
			scenario := testutil.CLIScenario{
				Name: tc.name, Args: "--workspace-id ws123 --category-id search-engines",
				Client: categoryClient(t, steps), WantOutput: tc.want,
			}
			testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, tc.command}, scenario)
		})
	}
}

func TestCategoryJSONDoesNotFetchOverrides(t *testing.T) {
	for _, command := range []string{"get", "describe"} {
		t.Run(command, func(t *testing.T) {
			testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, command}, testutil.CLIScenario{
				Name:           "JSON remains the category API response",
				Args:           "--workspace-id ws123 --category-id search-engines --json",
				Client:         categoryClient(t, []categoryRequest{{method: http.MethodGet, path: categoryPath, response: string(testutil.GenJSON(category))}}),
				WantOutput:     fstfmt.EncodeJSON(category),
				DontWantOutput: "override",
			})
		})
	}
}

func TestCategoryListDoesNotFetchOverrides(t *testing.T) {
	for _, flags := range []string{"", " --json"} {
		t.Run("list"+flags, func(t *testing.T) {
			want := "search-engines  Search Engines  allow"
			if flags != "" {
				want = fstfmt.EncodeJSON([]any{category})
			}
			testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, "list"}, testutil.CLIScenario{
				Name: "list has no override lookup",
				Args: "--workspace-id ws123" + flags,
				Client: categoryClient(t, []categoryRequest{{
					method: http.MethodGet, path: "api.example.com/bot-management/v1/workspaces/ws123/policy/categories",
					response: `{"data":[` + string(testutil.GenJSON(category)) + `]}`,
				}}),
				WantOutput: want, DontWantOutput: "override",
			})
		})
	}
}

func TestCategoryOverrideLookupError(t *testing.T) {
	testutil.RunCLIScenario(t, []string{root.CommandName, sub.CommandName, "get"}, testutil.CLIScenario{
		Name: "lookup failure does not report a misleading count",
		Args: "--workspace-id ws123 --category-id search-engines",
		Client: categoryClient(t, []categoryRequest{
			{method: http.MethodGet, path: categoryPath, response: string(testutil.GenJSON(category))},
			{method: http.MethodGet, path: categoryPath + "/bots", status: http.StatusNotFound, response: `{"title":"not found","status":404}`},
		}),
		WantError: "404 - Not Found", DontWantOutput: "Action:",
	})
}
