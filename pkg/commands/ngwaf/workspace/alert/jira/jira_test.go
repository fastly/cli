package jira_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	workspaceroot "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	alertroot "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/jira"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/jira"
)

const (
	alertID     = "890abcdef1234567890123ab"
	workspaceID = "nBw2ENWfOY1M2dpSwK1l5R"
	description = "TestJiraAlert"
)

var (
	host      = "example.atlassian.net"
	key       = "jira-api-key-123456"
	project   = "PROJ"
	username  = "user@example.com"
	issueType = "Task"
	jiraAlert = jira.Alert{
		ID:          alertID,
		Type:        "jira",
		Description: description,
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: jira.ResponseConfig{
			Host:      &host,
			Key:       &key,
			Project:   &project,
			Username:  &username,
			IssueType: &issueType,
		},
	}
)

func TestJiraAlertCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--host %s --key %s --project %s --username %s", host, key, project, username),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --host flag",
			Args:      fmt.Sprintf("--workspace-id %s --key %s --project %s --username %s", workspaceID, key, project, username),
			WantError: "error parsing arguments: required flag --host not provided",
		},
		{
			Name:      "validate missing --key flag",
			Args:      fmt.Sprintf("--workspace-id %s --host %s --project %s --username %s", workspaceID, host, project, username),
			WantError: "error parsing arguments: required flag --key not provided",
		},
		{
			Name:      "validate missing --project flag",
			Args:      fmt.Sprintf("--workspace-id %s --host %s --key %s --username %s", workspaceID, host, key, username),
			WantError: "error parsing arguments: required flag --project not provided",
		},
		{
			Name:      "validate missing --username flag",
			Args:      fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s", workspaceID, host, key, project),
			WantError: "error parsing arguments: required flag --username not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s --username %s", workspaceID, host, key, project, username),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", jiraAlert.Type, jiraAlert.ID, workspaceID),
		},
		{
			Name: "validate API success with description",
			Args: fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s --username %s --description %s", workspaceID, host, key, project, username, description),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", jiraAlert.Type, jiraAlert.ID, workspaceID),
		},
		{
			Name: "validate API success with issue-type",
			Args: fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s --username %s --issue-type %s", workspaceID, host, key, project, username, issueType),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", jiraAlert.Type, jiraAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s --username %s --json", workspaceID, host, key, project, username),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(jiraAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestJiraAlertList(t *testing.T) {
	alertsObject := jira.Alerts{
		Data: []jira.Alert{
			{
				ID:          "1a2b3c4d5e6f7890abcdef12",
				Type:        "jira",
				Description: "First Jira alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: jira.ResponseConfig{
					Host:      &host,
					Key:       &key,
					Project:   &project,
					Username:  &username,
					IssueType: &issueType,
				},
			},
			{
				ID:          "2b3c4d5e6f7890abcdef1234",
				Type:        "jira",
				Description: "Second Jira alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: jira.ResponseConfig{
					Host:      &host,
					Key:       &key,
					Project:   &project,
					Username:  &username,
					IssueType: &issueType,
				},
			},
		},
		Meta: jira.MetaAlerts{
			Total: 2,
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      "",
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
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
			Name: "validate API success (zero alerts)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(jira.Alerts{
							Data: []jira.Alert{},
						}))),
					},
				},
			},
			WantOutput: zeroListString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(alertsObject))),
					},
				},
			},
			WantOutput: listString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(alertsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(alertsObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestJiraAlertGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s", alertID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --alert-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --alert-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "This resource does not exist",
    							"status": 404
							}
						`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: alertString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --json", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(jiraAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(jiraAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestJiraAlertUpdate(t *testing.T) {
	updatedHost := "updated.atlassian.net"
	updatedKey := "updated-jira-key-456"
	updatedProject := "UPDT"
	updatedUsername := "updated@example.com"
	updatedIssueType := "Bug"
	updatedAlert := jira.Alert{
		ID:          alertID,
		Type:        "jira",
		Description: "Updated description",
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: jira.ResponseConfig{
			Host:      &updatedHost,
			Key:       &updatedKey,
			Project:   &updatedProject,
			Username:  &updatedUsername,
			IssueType: &updatedIssueType,
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s --host %s --key %s --project %s --username %s", alertID, host, key, project, username),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --alert-id flag",
			Args:      fmt.Sprintf("--workspace-id %s --host %s --key %s --project %s --username %s", workspaceID, host, key, project, username),
			WantError: "error parsing arguments: required flag --alert-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid --host updated.atlassian.net --key updated-jira-key-456 --project UPDT --username updated@example.com", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "This resource does not exist",
    							"status": 404
							}
						`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
		{
			Name: "validate API success with all config fields",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --host updated.atlassian.net --key updated-jira-key-456 --project UPDT --username updated@example.com --issue-type Bug", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,
							Status:     http.StatusText(http.StatusOK),
							Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(jiraAlert))),
						},
						{
							StatusCode: http.StatusOK,
							Status:     http.StatusText(http.StatusOK),
							Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
						},
					},
				},
			},
			WantOutput: fstfmt.Success("Updated '%s' alert '%s' (workspace-id: %s)", updatedAlert.Type, updatedAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --host updated.atlassian.net --key updated-jira-key-456 --project UPDT --username updated@example.com --json", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,
							Status:     http.StatusText(http.StatusOK),
							Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(jiraAlert))),
						},
						{
							StatusCode: http.StatusOK,
							Status:     http.StatusText(http.StatusOK),
							Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
						},
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatedAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestJiraAlertDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s", alertID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --alert-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --alert-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "This resource does not exist",
    							"status": 404
							}
						`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted alert '%s' (workspace-id: %s)", alertID, workspaceID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "delete"}, scenarios)
}

var alertString = strings.TrimSpace(`
ID: 890abcdef1234567890123ab
Type: jira
Description: TestJiraAlert
Created At: 2025-11-25T16:40:12Z
Created By: test@example.com
Config:
  Host: example.atlassian.net
  Username: user@example.com
  Project: PROJ
  Issue Type: Task
  Key: <redacted>
`)

var listString = strings.TrimSpace(`
ID                        Type  Description        Created At            Created By        Config
1a2b3c4d5e6f7890abcdef12  jira  First Jira alert   2025-11-25T16:40:12Z  test@example.com  Host: example.atlassian.net, Issue Type: Task, Key: <redacted>, Project: PROJ, Username: user@example.com
2b3c4d5e6f7890abcdef1234  jira  Second Jira alert  2025-11-25T16:40:12Z  test@example.com  Host: example.atlassian.net, Issue Type: Task, Key: <redacted>, Project: PROJ, Username: user@example.com
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Type  Description  Created At  Created By  Config
`) + "\n"
