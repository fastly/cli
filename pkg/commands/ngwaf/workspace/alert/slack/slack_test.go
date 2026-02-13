package slack_test

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
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/slack"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/slack"
)

const (
	alertID     = "1a2b3c4d5e6f7890abcdef12"
	workspaceID = "nBw2ENWfOY1M2dpSwK1l5R"
	description = "TestSlackAlert"
)

var (
	webhook    = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
	slackAlert = slack.Alert{
		ID:          alertID,
		Type:        "slack",
		Description: description,
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: slack.ResponseConfig{
			Webhook: &webhook,
		},
	}
)

func TestSlackAlertCreate(t *testing.T) {
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(slackAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", slackAlert.Type, slackAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --webhook %s --json", workspaceID, webhook),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(slackAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(slackAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestSlackAlertList(t *testing.T) {
	alertsObject := slack.Alerts{
		Data: []slack.Alert{
			{
				ID:          "1a2b3c4d5e6f7890abcdef12",
				Type:        "slack",
				Description: "First slack alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: slack.ResponseConfig{
					Webhook: &webhook,
				},
			},
			{
				ID:          "2b3c4d5e6f7890abcdef1234",
				Type:        "slack",
				Description: "Second slack alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: slack.ResponseConfig{
					Webhook: &webhook,
				},
			},
		},
		Meta: slack.MetaAlerts{
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
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(slack.Alerts{
							Data: []slack.Alert{},
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

func TestSlackAlertGet(t *testing.T) {
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(slackAlert)))),
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(slackAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(slackAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestSlackAlertUpdate(t *testing.T) {
	updatedWebhook := "https://hooks.slack.com/services/updated"
	updatedAlert := slack.Alert{
		ID:          alertID,
		Type:        "slack",
		Description: "Updated description",
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: slack.ResponseConfig{
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
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid --webhook https://hooks.slack.com/services/updated", workspaceID),
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
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --webhook https://hooks.slack.com/services/updated", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(slackAlert))),
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
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --webhook https://hooks.slack.com/services/updated --json", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MultiResponseRoundTripper{
					Responses: []*http.Response{
						{
							StatusCode: http.StatusOK,

							Status: http.StatusText(http.StatusOK),

							Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(slackAlert))),
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

func TestSlackAlertDelete(t *testing.T) {
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
ID: 1a2b3c4d5e6f7890abcdef12
Type: slack
Description: TestSlackAlert
Created At: 2025-11-25T16:40:12Z
Created By: test@example.com
Config:
  Webhook: <redacted>
`)

var listString = strings.TrimSpace(`
ID                        Type   Description         Created At            Created By        Config
1a2b3c4d5e6f7890abcdef12  slack  First slack alert   2025-11-25T16:40:12Z  test@example.com  Webhook: <redacted>
2b3c4d5e6f7890abcdef1234  slack  Second slack alert  2025-11-25T16:40:12Z  test@example.com  Webhook: <redacted>
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Type  Description  Created At  Created By  Config
`) + "\n"
