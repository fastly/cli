package workspace_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/workspace"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

var workspace = workspaces.Workspace{
	ID:             "ws123",
	Name:           "Production",
	Description:    "Main site",
	ProtectionMode: "log",
	Services:       []string{"svcA", "svcB"},
	CreatedAt:      "2026-01-01T00:00:00Z",
	UpdatedAt:      "2026-02-01T00:00:00Z",
}

func TestWorkspaceList(t *testing.T) {
	page := workspaces.Workspaces{
		Data: []workspaces.Workspace{
			workspace,
			{ID: "ws456", Name: "Staging", ProtectionMode: "off", UpdatedAt: "2026-03-01T00:00:00Z"},
		},
	}

	unfiltered := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(page)}
	filtered := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(page)}

	scenarios := []testutil.CLIScenario{
		{
			Name: "API error is surfaced",
			Client: (&testutil.RequestRecorder{
				Status: http.StatusInternalServerError,
				Body:   []byte(`{"title":"boom","status":500}`),
			}).Client(),
			WantError: "500 - Internal Server Error",
		},
		{
			Name:   "lists workspaces as a table without a service filter",
			Client: unfiltered.Client(),
			WantOutput: strings.TrimSpace(`
ID     Name        Protection Mode  Services    Updated At
ws123  Production  log              svcA, svcB  2026-02-01T00:00:00Z
ws456  Staging     off                          2026-03-01T00:00:00Z
`) + "\n",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if unfiltered.Path != "/bot-management/v1/workspaces" {
					t.Errorf("unexpected path: %s", unfiltered.Path)
				}
				if unfiltered.Query.Has("service_id") {
					t.Errorf("service_id should not be sent when --service-id is omitted, got %q", unfiltered.Query.Get("service_id"))
				}
			},
		},
		{
			Name:        "--service-id filters by service",
			Args:        "--service-id svcA",
			Client:      filtered.Client(),
			WantOutputs: []string{"ws123"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if got := filtered.Query.Get("service_id"); got != "svcA" {
					t.Errorf("want service_id=svcA, got %q", got)
				}
			},
		},
		{
			Name:       "--json emits the workspaces",
			Args:       "--json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(page)}).Client(),
			WantOutput: fstfmt.EncodeJSON(page.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestWorkspaceDescribe(t *testing.T) {
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(workspace)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:   "describes the workspace",
			Args:   "--workspace-id ws123",
			Client: rec.Client(),
			WantOutput: strings.TrimSpace(`
ID: ws123
Name: Production
Description: Main site
Protection Mode: log
Services: svcA, svcB
Created At: 2026-01-01T00:00:00Z
Updated At: 2026-02-01T00:00:00Z
`),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodGet || rec.Path != "/bot-management/v1/workspaces/ws123" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
			},
		},
		{
			Name:       "--json emits the workspace",
			Args:       "--workspace-id ws123 --json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(workspace)}).Client(),
			WantOutput: fstfmt.EncodeJSON(workspace),
		},
		{
			Name: "API error is surfaced",
			Args: "--workspace-id missing",
			Client: (&testutil.RequestRecorder{
				Status: http.StatusNotFound,
				Body:   []byte(`{"title":"not found","status":404}`),
			}).Client(),
			WantError: "404 - Not Found",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestWorkspaceUpdate(t *testing.T) {
	updated := workspace
	updated.ProtectionMode = "block"
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(updated)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			Args:      "--name foo",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:            "at least one field must be provided",
			Args:            "--workspace-id ws123",
			WantError:       "no workspace fields to update",
			WantRemediation: "--protection-mode",
		},
		{
			Name:      "--protection-mode rejects unknown values",
			Args:      "--workspace-id ws123 --protection-mode challenge",
			WantError: "enum value must be one of off,log,block, got 'challenge'",
		},
		{
			Name:       "sends only the provided fields",
			Args:       "--workspace-id ws123 --protection-mode block",
			Client:     rec.Client(),
			WantOutput: fstfmt.Success("Updated Bot Management workspace 'Production' (workspace-id: ws123, protection-mode: block)"),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodPatch || rec.Path != "/bot-management/v1/workspaces/ws123" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
				var body map[string]any
				if err := json.Unmarshal(rec.RequestBody, &body); err != nil {
					t.Fatalf("request body is not JSON: %v (%s)", err, rec.RequestBody)
				}
				if len(body) != 1 || body["protection_mode"] != "block" {
					t.Errorf("want body {\"protection_mode\":\"block\"}, got %s", rec.RequestBody)
				}
			},
		},
		{
			Name:       "--json emits the updated workspace",
			Args:       "--workspace-id ws123 --name Production --json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(updated)}).Client(),
			WantOutput: fstfmt.EncodeJSON(updated),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
