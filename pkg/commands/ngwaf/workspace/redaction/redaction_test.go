package redaction_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	workspace "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/redaction"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/redactions"
)

const (
	redactionField = "password"
	redactionID    = "someID"
	redactionType  = "request"
	workspaceID    = "workspaceID"
)

var redaction = redactions.Redaction{
	CreatedAt:   testutil.Date,
	Field:       redactionField,
	RedactionID: redactionID,
	Type:        redactionType,
}

func TestRedactionCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --field flag",
			Args:      fmt.Sprintf("--type %s --workspace-id %s", redactionType, workspaceID),
			WantError: "error parsing arguments: required flag --field not provided",
		},
		{
			Name:      "validate missing --type flag",
			Args:      fmt.Sprintf("--field %s --workspace-id %s", redactionField, workspaceID),
			WantError: "error parsing arguments: required flag --type not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--field %s --type %s", redactionField, redactionType),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--field %s --type %s --workspace-id %s", redactionField, redactionType, workspaceID),
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
			Args: fmt.Sprintf("--field %s --type %s --workspace-id %s", redactionField, redactionType, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(redaction)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created redaction '%s' (field: %s, type: %s)", redactionID, redactionField, redactionType),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--field %s --type %s --workspace-id %s --json", redactionField, redactionType, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(redaction))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(redaction),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestRedactionDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --redaction-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --redaction-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--redaction-id %s", redactionID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Redaction ID",
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
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted redaction (id: %s)", redactionID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s --json", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, redactionID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestRedactionRetrieve(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --redaction-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --redaction-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--redaction-id %s", redactionID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Redaction ID",
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
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(redaction)))),
					},
				},
			},
			WantOutput: redactionString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s --json", redactionID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(redaction)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(redaction),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "retrieve"}, scenarios)
}

func TestRedactionList(t *testing.T) {
	redactionsObject := redactions.Redactions{
		Data: []redactions.Redaction{
			{
				CreatedAt:   testutil.Date,
				Field:       redactionField,
				RedactionID: redactionID,
				Type:        redactionType,
			},
			{
				CreatedAt:   testutil.Date,
				Field:       "username",
				RedactionID: redactionID + "2",
				Type:        redactionType,
			},
		},
		Meta: redactions.MetaRedactions{},
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
			Name: "validate API success (zero redactions)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(redactions.Redactions{
							Data: []redactions.Redaction{},
							Meta: redactions.MetaRedactions{},
						}))),
					},
				},
			},
			WantOutput: zeroListRedactionString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(redactionsObject))),
					},
				},
			},
			WantOutput: listRedactionString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(redactionsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(redactionsObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestRedactionUpdate(t *testing.T) {
	redactionsObject := redactions.Redaction{
		CreatedAt:   testutil.Date,
		Field:       redactionField,
		RedactionID: redactionID,
		Type:        redactionType,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --redaction-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --redaction-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--redaction-id %s", redactionID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s --field %s --type %s", redactionID, workspaceID, redactionField, redactionType),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(redactionsObject))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated redaction '%s' (field: %s, type: %s)", redactionID, redactionField, redactionType),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--redaction-id %s --workspace-id %s --field %s --type %s --json", redactionID, workspaceID, redactionField, redactionType),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(redaction))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(redaction),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "update"}, scenarios)
}

var listRedactionString = strings.TrimSpace(`
Field     ID       Type     Created At
password  someID   request  2021-06-15 23:00:00 +0000 UTC
username  someID2  request  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListRedactionString = strings.TrimSpace(`
Field  ID  Type  Created At
`) + "\n"

var redactionString = strings.TrimSpace(`
Field: password
ID: someID
Type: request
Created At: 2021-06-15 23:00:00 +0000 UTC
`)
