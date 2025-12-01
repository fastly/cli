package customsignal_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	sub2 "github.com/fastly/cli/pkg/commands/ngwaf/workspace/customsignal"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"
)

const (
	customSignalDescription = "NGWAFCLICustomSignal"
	customSignalID          = "someID"
	customSignalName        = "CLICustomSignalName"
	workspaceID             = "WorkspaceID"
)

var customSignal = signals.Signal{
	CreatedAt:   testutil.Date,
	Description: customSignalDescription,
	Name:        customSignalName,
	SignalID:    customSignalID,
	Scope: signals.Scope{
		Type:      string(scope.ScopeTypeWorkspace),
		AppliesTo: []string{workspaceID},
	},
}

func TestCustomSignalCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--description %s --workspace-id %s", customSignalDescription, workspaceID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--description %s --name %s", customSignalDescription, customSignalName),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--description %s --name %s --workspace-id %s", customSignalDescription, customSignalName, workspaceID),
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
			Args: fmt.Sprintf("--description %s --name %s --workspace-id %s", customSignalDescription, customSignalName, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(customSignal)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created workspace-level custom signal '%s' (signal-id: %s)", customSignalName, customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--description %s --name %s --workspace-id %s --json", customSignalDescription, customSignalName, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignal))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(customSignal),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, sub2.CommandName, "create"}, scenarios)
}

func TestCustomSignalDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --signal-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--signal-id %s", customSignalID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--signal-id bar --workspace-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid signal ID",
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
			Args: fmt.Sprintf("--signal-id %s --workspace-id %s", customSignalID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted workspace-level custom signal (id: %s)", customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--signal-id %s --workspace-id %s --json", customSignalID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, customSignalID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, sub2.CommandName, "delete"}, scenarios)
}

func TestCustomSignalGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --signal-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--signal-id %s", customSignalID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--signal-id baz --workspace-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Custom Signal ID",
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
			Args: fmt.Sprintf("--signal-id %s --workspace-id %s", customSignalID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(customSignal)))),
					},
				},
			},
			WantOutput: customSignalString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--signal-id %s --workspace-id %s --json", customSignalID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(customSignal)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(customSignal),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, sub2.CommandName, "get"}, scenarios)
}

func TestCustomSignalList(t *testing.T) {
	customSignalsObject := signals.Signals{
		Data: []signals.Signal{
			{
				CreatedAt:   testutil.Date,
				Description: customSignalDescription,
				Name:        customSignalName,
				SignalID:    customSignalID,
				Scope: signals.Scope{
					Type:      string(scope.ScopeTypeWorkspace),
					AppliesTo: []string{workspaceID},
				},
			},
			{
				CreatedAt:   testutil.Date,
				Description: customSignalDescription,
				Name:        customSignalName + "2",
				SignalID:    customSignalID + "2",
				Scope: signals.Scope{
					Type:      string(scope.ScopeTypeWorkspace),
					AppliesTo: []string{workspaceID},
				},
			},
		},
		Meta: signals.MetaSignals{},
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
			Name: "validate API success (zero workspace-level custom signals)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(signals.Signals{
							Data: []signals.Signal{},
							Meta: signals.MetaSignals{},
						}))),
					},
				},
			},
			WantOutput: zeroListCustomSignalsString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignalsObject))),
					},
				},
			},
			WantOutput: listCustomSignalsString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignalsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(customSignalsObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, sub2.CommandName, "list"}, scenarios)
}

func TestCustomSignalUpdate(t *testing.T) {
	customSignalObject := signals.Signal{
		CreatedAt:   testutil.Date,
		Description: customSignalDescription,
		Name:        customSignalName,
		SignalID:    customSignalID,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --signal-id flag",
			Args:      fmt.Sprintf("--description %s --workspace-id %s", customSignalDescription+"2", workspaceID),
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--description %s --signal-id %s", customSignalDescription+"2", customSignalID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --description flag",
			Args:      fmt.Sprintf("--workspace-id %s --signal-id %s", workspaceID, customSignalID),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--signal-id %s --description %s --workspace-id %s", customSignalID, customSignalDescription, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignalObject))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated workspace-level signal '%s' (signal-id: %s)", customSignalName, customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--signal-id %s --description %s --workspace-id %s --json", customSignalID, customSignalDescription, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignal))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(customSignal),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, sub2.CommandName, "update"}, scenarios)
}

var listCustomSignalsString = strings.TrimSpace(`
ID       Name                  Description           Scope      Updated At                     Created At
someID   CLICustomSignalName   NGWAFCLICustomSignal  workspace  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
someID2  CLICustomSignalName2  NGWAFCLICustomSignal  workspace  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListCustomSignalsString = strings.TrimSpace(`
ID  Name  Description  Scope  Updated At  Created At
`) + "\n"

var customSignalString = strings.TrimSpace(`
ID: someID
Name: CLICustomSignalName
Description: NGWAFCLICustomSignal
Scope: workspace
Updated (UTC): 0001-01-01 00:00
Created (UTC): 2021-06-15 23:00
`)
