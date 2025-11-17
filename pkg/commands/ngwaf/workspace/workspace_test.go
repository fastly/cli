package workspace_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces"
)

const (
	workspaceDescription     = "NGWAFCLIWorkspace"
	workspaceClientIPHeaders = "these:are:headers"
	workspaceID              = "someID"
	workspaceMode            = "log"
	workspaceName            = "CLIWorkspace"
)

var workspace = workspaces.Workspace{
	AttackSignalThresholds: workspaces.AttackSignalThresholds{
		Immediate:  false,
		OneMinute:  0,
		TenMinutes: 0,
		OneHour:    0,
	},
	ClientIPHeaders: []string{"these", "are", "headers"},
	CreatedAt:       testutil.Date,
	Description:     workspaceDescription,
	Mode:            workspaceMode,
	Name:            workspaceName,
	WorkspaceID:     workspaceID,
}

func TestWorkspacesCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --description flag",
			Args:      fmt.Sprintf("--blockingMode %s --name %s", workspaceMode, workspaceName),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name:      "validate missing --blockingMode flag",
			Args:      fmt.Sprintf("--description %s --name %s", workspaceDescription, workspaceName),
			WantError: "error parsing arguments: required flag --blockingMode not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--description %s --blockingMode %s", workspaceDescription, workspaceMode),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--description %s --name %s --blockingMode %s", workspaceDescription, workspaceName, "invalidMode"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--description %s --name %s --blockingMode %s --clientIPHeaders %s", workspaceDescription, workspaceName, workspaceMode, workspaceClientIPHeaders),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(workspace)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created workspace '%s' (workspace-id: %s)", workspaceName, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--description %s --name %s --blockingMode %s --clientIPHeaders %s --json", workspaceDescription, workspaceName, workspaceMode, workspaceClientIPHeaders),
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
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestWorkspaceDelete(t *testing.T) {
	const workspaceID = "workspaceID"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--workspace-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Workspace ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted workspace (id: %s)", workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, workspaceID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestWorkspaceGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--workspace-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Workspace ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(workspace)))),
					},
				},
			},
			WantOutput: workspaceString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(workspace)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(workspace),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestWorkspaceList(t *testing.T) {
	workspacesObject := workspaces.Workspaces{
		Data: []workspaces.Workspace{
			{
				CreatedAt:   testutil.Date,
				Description: workspaceDescription,
				Mode:        workspaceMode,
				Name:        workspaceName,
				WorkspaceID: workspaceID,
			},
			{
				CreatedAt:   testutil.Date,
				Description: workspaceDescription,
				Mode:        workspaceMode,
				Name:        workspaceName + "2",
				WorkspaceID: workspaceID + "2",
			},
		},
		Meta: workspaces.MetaWorkspaces{},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate internal server error",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success (zero workspaces)",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(workspaces.Workspaces{
							Data: []workspaces.Workspace{},
							Meta: workspaces.MetaWorkspaces{},
						}))),
					},
				},
			},
			WantOutput: zeroListWorkspaceString,
		},
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(workspacesObject))),
					},
				},
			},
			WantOutput: listWorkspaceString,
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(workspacesObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(workspacesObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

var listWorkspaceString = strings.TrimSpace(`
ID       Name           Description        Mode  Created At
someID   CLIWorkspace   NGWAFCLIWorkspace  log   2021-06-15 23:00:00 +0000 UTC
someID2  CLIWorkspace2  NGWAFCLIWorkspace  log   2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListWorkspaceString = strings.TrimSpace(`
ID  Name  Description  Mode  Created At
`) + "\n"

var workspaceString = strings.TrimSpace(`
ID: someID
Name: CLIWorkspace
Description: NGWAFCLIWorkspace
Mode: log
Attack Signal Thresholds:
	Immediate: false
	One Minute: 0
	Ten Minutes: 0
	One Hour: 0
Client IP Headers: these, are, headers
Updated (UTC): 0001-01-01 00:00
`)
