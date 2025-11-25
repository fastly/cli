package datadog_test

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
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/datadog"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

const (
	alertID     = "7890abcdef12345678901234"
	workspaceID = "nBw2ENWfOY1M2dpSwK1l5R"
	description = "Test Datadog alert"
)

var (
	key          = "a1b2c3d4e5f67890abcdef1234567890"
	site         = "datadoghq.com"
	datadogAlert = datadog.Alert{
		ID:          alertID,
		Type:        "datadog",
		Description: description,
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: datadog.ResponseConfig{
			Key:  &key,
			Site: &site,
		},
	}
)

func TestDatadogAlertCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--key %s --site %s", key, site),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --key flag",
			Args:      fmt.Sprintf("--workspace-id %s --site %s", workspaceID, site),
			WantError: "error parsing arguments: required flag --key not provided",
		},
		{
			Name:      "validate missing --site flag",
			Args:      fmt.Sprintf("--workspace-id %s --key %s", workspaceID, key),
			WantError: "error parsing arguments: required flag --site not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --key %s --site %s", workspaceID, key, site),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(datadogAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", datadogAlert.Type, datadogAlert.ID, workspaceID),
		},
		{
			Name: "validate API success with description",
			Args: fmt.Sprintf("--workspace-id %s --key %s --site %s --description %s", workspaceID, key, site, description),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(datadogAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", datadogAlert.Type, datadogAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --key %s --site %s --json", workspaceID, key, site),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(datadogAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(datadogAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDatadogAlertList(t *testing.T) {
	alertsObject := datadog.Alerts{
		Data: []datadog.Alert{
			{
				ID:          "1a2b3c4d5e6f7890abcdef12",
				Type:        "datadog",
				Description: "First Datadog alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: datadog.ResponseConfig{
					Key:  &key,
					Site: &site,
				},
			},
			{
				ID:          "2b3c4d5e6f7890abcdef1234",
				Type:        "datadog",
				Description: "Second Datadog alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: datadog.ResponseConfig{
					Key:  &key,
					Site: &site,
				},
			},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
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
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(datadog.Alerts{
							Data: []datadog.Alert{},
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

func TestDatadogAlertGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s", alertID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(datadogAlert)))),
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(datadogAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(datadogAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestDatadogAlertUpdate(t *testing.T) {
	updatedKey := "updated-key-9876543210"
	updatedSite := "datadoghq.eu"
	updatedAlert := datadog.Alert{
		ID:          alertID,
		Type:        "datadog",
		Description: "Updated description",
		Config: datadog.ResponseConfig{
			Key:  &updatedKey,
			Site: &updatedSite,
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s", alertID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --alert-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --alert-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid --key updated-key-9876543210", workspaceID),
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
			Name: "validate API success with key and site",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --key updated-key-9876543210 --site datadoghq.eu", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated '%s' alert '%s' (workspace-id: %s)", updatedAlert.Type, updatedAlert.ID, workspaceID),
		},
		{
			Name: "validate API success with description",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --description \"Updated description\"", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated '%s' alert '%s' (workspace-id: %s)", updatedAlert.Type, updatedAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --key updated-key-9876543210 --site datadoghq.eu --json", workspaceID, alertID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedAlert))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatedAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestDatadogAlertDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--alert-id %s", alertID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
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
ID: 7890abcdef12345678901234
Type: datadog
Description: Test Datadog alert
Created At: 2025-11-25T16:40:12Z
Created By: test@example.com
Config:
  Key: <redacted>
  Site: datadoghq.com
`)

var listString = strings.TrimSpace(`
ID         Type     Description           Created At            Created By        Config
1a2b3c4d5e6f7890abcdef12  datadog  First Datadog alert   2025-11-25T16:40:12Z  test@example.com  Site: datadoghq.com, Key: <redacted>
2b3c4d5e6f7890abcdef1234  datadog  Second Datadog alert  2025-11-25T16:40:12Z  test@example.com  Site: datadoghq.com, Key: <redacted>
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Type  Description  Created At  Created By  Config
`) + "\n"
