package microsoftteams_test

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
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/microsoftteams"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
)

const (
	alertID     = "3c4d5e6f7890abcdef123456"
	workspaceID = "nBw2ENWfOY1M2dpSwK1l5R"
	description = "TestMicrosoftTeamsAlert"
)

var (
	webhook    = "https://outlook.office.com/webhook/example"
	teamsAlert = microsoftteams.Alert{
		ID:          alertID,
		Type:        "microsoftteams",
		Description: description,
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: microsoftteams.ResponseConfig{
			Webhook: &webhook,
		},
	}
)

func TestMicrosoftTeamsAlertCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--webhook %s", webhook),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --webhook flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --webhook not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --webhook %s", workspaceID, webhook),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(teamsAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", teamsAlert.Type, teamsAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --webhook %s --json", workspaceID, webhook),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(teamsAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(teamsAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestMicrosoftTeamsAlertList(t *testing.T) {
	alertsObject := microsoftteams.Alerts{
		Data: []microsoftteams.Alert{
			{
				ID:          "1a2b3c4d5e6f7890abcdef12",
				Type:        "microsoftteams",
				Description: "First Microsoft Teams alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: microsoftteams.ResponseConfig{
					Webhook: &webhook,
				},
			},
			{
				ID:          "2b3c4d5e6f7890abcdef1234",
				Type:        "microsoftteams",
				Description: "Second Microsoft Teams alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: microsoftteams.ResponseConfig{
					Webhook: &webhook,
				},
			},
		},
		Meta: microsoftteams.MetaAlerts{
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
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(microsoftteams.Alerts{
							Data: []microsoftteams.Alert{},
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

func TestMicrosoftTeamsAlertGet(t *testing.T) {
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(teamsAlert)))),
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(teamsAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(teamsAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestMicrosoftTeamsAlertUpdate(t *testing.T) {
	updatedWebhook := "https://outlook.office.com/webhook/updated"
	updatedAlert := microsoftteams.Alert{
		ID:          alertID,
		Type:        "microsoftteams",
		Description: "Updated description",
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: microsoftteams.ResponseConfig{
			Webhook: &updatedWebhook,
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s --webhook %s", alertID, webhook),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name:      "validate missing --alert-id flag",
			Args:      fmt.Sprintf("--workspace-id %s --webhook %s", workspaceID, webhook),
			WantError: "error parsing arguments: required flag --alert-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid --webhook https://outlook.office.com/webhook/updated", workspaceID),
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
			Name: "validate API success with webhook",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --webhook https://outlook.office.com/webhook/updated", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(teamsAlert))),
						},

						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
						},
					},
				},
			},
			WantOutput: fstfmt.Success("Updated '%s' alert '%s' (workspace-id: %s)", updatedAlert.Type, updatedAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --webhook https://outlook.office.com/webhook/updated --json", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(teamsAlert))),
						},

						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
						},
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatedAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestMicrosoftTeamsAlertDelete(t *testing.T) {
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
ID: 3c4d5e6f7890abcdef123456
Type: microsoftteams
Description: TestMicrosoftTeamsAlert
Created At: 2025-11-25T16:40:12Z
Created By: test@example.com
Config:
  Webhook: <redacted>
`)

var listString = strings.TrimSpace(`
ID                        Type            Description                   Created At            Created By        Config
1a2b3c4d5e6f7890abcdef12  microsoftteams  First Microsoft Teams alert   2025-11-25T16:40:12Z  test@example.com  Webhook: <redacted>
2b3c4d5e6f7890abcdef1234  microsoftteams  Second Microsoft Teams alert  2025-11-25T16:40:12Z  test@example.com  Webhook: <redacted>
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Type  Description  Created At  Created By  Config
`) + "\n"
