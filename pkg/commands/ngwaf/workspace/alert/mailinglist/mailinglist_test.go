package mailinglist_test

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
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/mailinglist"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/mailinglist"
)

const (
	alertID     = "2b3c4d5e6f7890abcdef1234"
	workspaceID = "nBw2ENWfOY1M2dpSwK1l5R"
	description = "Test mailing list alert"
)

var (
	address          = "alerts@example.com"
	mailinglistAlert = mailinglist.Alert{
		ID:          alertID,
		Type:        "mailingList",
		Description: description,
		CreatedAt:   "2025-11-25T16:40:12Z",
		CreatedBy:   "test@example.com",
		Config: mailinglist.ResponseConfig{
			Address: &address,
		},
	}
)

func TestMailingListAlertCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--address %s", address),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --address flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --address not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --address %s", workspaceID, address),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(mailinglistAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", mailinglistAlert.Type, mailinglistAlert.ID, workspaceID),
		},
		{
			Name: "validate API success with description",
			Args: fmt.Sprintf("--workspace-id %s --address %s --description %s", workspaceID, address, description),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(mailinglistAlert)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created a '%s' alert '%s' (workspace-id: %s)", mailinglistAlert.Type, mailinglistAlert.ID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --address %s --json", workspaceID, address),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(mailinglistAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(mailinglistAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestMailingListAlertList(t *testing.T) {
	alertAddress1 := "alerts1@example.com"
	alertAddress2 := "alerts2@example.com"
	alertsObject := mailinglist.Alerts{
		Data: []mailinglist.Alert{
			{
				ID:          "1a2b3c4d5e6f7890abcdef12",
				Type:        "mailingList",
				Description: "First mailing list alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: mailinglist.ResponseConfig{
					Address: &alertAddress1,
				},
			},
			{
				ID:          "2b3c4d5e6f7890abcdef1234",
				Type:        "mailingList",
				Description: "Second mailing list alert",
				CreatedAt:   "2025-11-25T16:40:12Z",
				CreatedBy:   "test@example.com",
				Config: mailinglist.ResponseConfig{
					Address: &alertAddress2,
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
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(mailinglist.Alerts{
							Data: []mailinglist.Alert{},
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

func TestMailingListAlertGet(t *testing.T) {
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(mailinglistAlert)))),
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
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(mailinglistAlert)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(mailinglistAlert),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspaceroot.CommandName, alertroot.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestMailingListAlertUpdate(t *testing.T) {
	updatedAddress := "updated@example.com"
	updatedAlert := mailinglist.Alert{
		ID:          alertID,
		Type:        "mailingList",
		Description: "Updated description",
		Config: mailinglist.ResponseConfig{
			Address: &updatedAddress,
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
			Args: fmt.Sprintf("--workspace-id %s --alert-id invalid --address updated@example.com", workspaceID),
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
			Name: "validate API success with address",
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --address updated@example.com", workspaceID, alertID),
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
			Args: fmt.Sprintf("--workspace-id %s --alert-id %s --address updated@example.com --json", workspaceID, alertID),
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

func TestMailingListAlertDelete(t *testing.T) {
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
ID: 2b3c4d5e6f7890abcdef1234
Type: mailingList
Description: Test mailing list alert
Created At: 2025-11-25T16:40:12Z
Created By: test@example.com
Config:
  Address: alerts@example.com
`)

var listString = strings.TrimSpace(`
ID         Type         Description                Created At            Created By        Config
1a2b3c4d5e6f7890abcdef12  mailingList  First mailing list alert   2025-11-25T16:40:12Z  test@example.com  Address: alerts1@example.com
2b3c4d5e6f7890abcdef1234  mailingList  Second mailing list alert  2025-11-25T16:40:12Z  test@example.com  Address: alerts2@example.com
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Type  Description  Created At  Created By  Config
`) + "\n"
