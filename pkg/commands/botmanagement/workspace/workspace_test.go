package workspace_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/workspace"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
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

	scenarios := []testutil.CLIScenario{
		{
			Name: "API error is surfaced",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"title":"boom","status":500}`))),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "lists workspaces as a table",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(page))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID     Name        Protection Mode  Services    Updated At
ws123  Production  log              svcA, svcB  2026-02-01T00:00:00Z
ws456  Staging     off                          2026-03-01T00:00:00Z
`) + "\n",
		},
		{
			Name: "accepts --service-id",
			Args: "--service-id svcA",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(page))),
					},
				},
			},
			WantOutputs: []string{"ws123"},
		},
		{
			Name: "--json emits the workspaces",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(page))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(page.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestWorkspaceDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "describes the workspace",
			Args: "--workspace-id ws123",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(workspace))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID: ws123
Name: Production
Description: Main site
Protection Mode: log
Services: svcA, svcB
Updated At: 2026-02-01T00:00:00Z
`),
		},
		{
			Name: "--json emits the workspace",
			Args: "--workspace-id ws123 --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(workspace))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(workspace),
		},
		{
			Name: "API error is surfaced",
			Args: "--workspace-id missing",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"title":"not found","status":404}`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestWorkspaceUpdate(t *testing.T) {
	updated := workspace
	updated.ProtectionMode = "block"

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
			Name:            "--name rejects values over 255 characters",
			Args:            "--workspace-id ws123 --name " + strings.Repeat("a", 256),
			WantError:       "error parsing arguments: --name must be between 1 and 255 characters, got 256",
			WantRemediation: "Provide a value for --name that is between 1 and 255 characters long.",
		},
		{
			Name:            "--name rejects an empty value",
			Args:            "--workspace-id ws123 --name=",
			WantError:       "error parsing arguments: --name must be between 1 and 255 characters, got 0",
			WantRemediation: "Provide a value for --name that is between 1 and 255 characters long.",
		},
		{
			Name:            "--description rejects values over 1000 characters",
			Args:            "--workspace-id ws123 --description " + strings.Repeat("a", 1001),
			WantError:       "error parsing arguments: --description must be between 0 and 1000 characters, got 1001",
			WantRemediation: "Provide a value for --description that is between 0 and 1000 characters long.",
		},
		{
			Name: "an empty --description counts as a field to update",
			Args: "--workspace-id ws123 --description=",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated Bot Management workspace 'Production' (workspace-id: ws123, protection-mode: block)"),
		},
		{
			Name: "prints the updated workspace",
			Args: "--workspace-id ws123 --protection-mode block",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated Bot Management workspace 'Production' (workspace-id: ws123, protection-mode: block)"),
		},
		{
			Name: "--json emits the updated workspace",
			Args: "--workspace-id ws123 --name Production --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updated),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
