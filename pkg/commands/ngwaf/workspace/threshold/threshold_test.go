package threshold_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	workspace "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/threshold"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/thresholds"
)

const (
	thresholdAction     = "block"
	thresholdDuration   = 86400
	thresholdEnabled    = true
	thresholdID         = "thresholdID"
	thresholdInterval   = 3600
	thresholdLimit      = 10
	thresholdName       = "Test_Threshold"
	thresholdSignal     = "test-signal"
	thresholdDontNotify = false
	workspaceID         = "workspaceID"
)

var threshold = thresholds.Threshold{
	Action:      thresholdAction,
	CreatedAt:   testutil.Date,
	DontNotify:  thresholdDontNotify,
	Duration:    thresholdDuration,
	Enabled:     thresholdEnabled,
	Interval:    thresholdInterval,
	Limit:       thresholdLimit,
	Name:        thresholdName,
	Signal:      thresholdSignal,
	ThresholdID: thresholdID,
}

func TestThresholdCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --action flag",
			Args:      fmt.Sprintf("--name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --action not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--action %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --signal flag",
			Args:      fmt.Sprintf("--action %s --name %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --signal not provided",
		},
		{
			Name:      "validate missing --do-not-notify flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --do-not-notify not provided",
		},
		{
			Name:      "validate missing --duration flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --duration not provided",
		},
		{
			Name:      "validate missing --enabled flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdInterval, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --enabled not provided",
		},
		{
			Name:      "validate missing --interval flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdLimit, workspaceID),
			WantError: "error parsing arguments: required flag --interval not provided",
		},
		{
			Name:      "validate missing --limit flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, workspaceID),
			WantError: "error parsing arguments: required flag --limit not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
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
			Args: fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(threshold)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created threshold '%s' for workspace '%s'", thresholdID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --workspace-id %s --json", thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(threshold))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(threshold),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestThresholdDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --threshold-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --threshold-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--threshold-id %s", thresholdID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Threshold ID",
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
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted threshold (id: %s)", thresholdID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s --json", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, thresholdID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestThresholdGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --threshold-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --threshold-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--threshold-id %s", thresholdID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Threshold ID",
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
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(threshold)))),
					},
				},
			},
			WantOutput: thresholdString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s --json", thresholdID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(threshold)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(threshold),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestThresholdList(t *testing.T) {
	thresholdsObject := thresholds.Thresholds{
		Data: []thresholds.Threshold{
			{
				Action:      thresholdAction,
				CreatedAt:   testutil.Date,
				DontNotify:  thresholdDontNotify,
				Duration:    thresholdDuration,
				Enabled:     thresholdEnabled,
				Interval:    thresholdInterval,
				Limit:       thresholdLimit,
				Name:        thresholdName,
				Signal:      thresholdSignal,
				ThresholdID: thresholdID,
			},
			{
				Action:      "log",
				CreatedAt:   testutil.Date,
				DontNotify:  true,
				Duration:    43200,
				Enabled:     false,
				Interval:    600,
				Limit:       20,
				Name:        "Test_Threshold_2",
				Signal:      "test-signal-2",
				ThresholdID: thresholdID + "2",
			},
		},
		Meta: thresholds.MetaThresholds{},
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
			Name: "validate API success (zero thresholds)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(thresholds.Thresholds{
							Data: []thresholds.Threshold{},
							Meta: thresholds.MetaThresholds{},
						}))),
					},
				},
			},
			WantOutput: zeroListThresholdString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(thresholdsObject))),
					},
				},
			},
			WantOutput: listThresholdString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(thresholdsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(thresholdsObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestThresholdUpdate(t *testing.T) {
	thresholdsObject := thresholds.Threshold{
		Action:      thresholdAction,
		CreatedAt:   testutil.Date,
		DontNotify:  thresholdDontNotify,
		Duration:    thresholdDuration,
		Enabled:     thresholdEnabled,
		Interval:    thresholdInterval,
		Limit:       thresholdLimit,
		Name:        thresholdName,
		Signal:      thresholdSignal,
		ThresholdID: thresholdID,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --threshold-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --threshold-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--threshold-id %s", thresholdID),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s --action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d", thresholdID, workspaceID, thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(thresholdsObject))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated threshold '%s' for workspace '%s'", thresholdID, workspaceID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--threshold-id %s --workspace-id %s --action %s --name %s --signal %s --do-not-notify=%t --duration %d --enabled=%t --interval %d --limit %d --json", thresholdID, workspaceID, thresholdAction, thresholdName, thresholdSignal, thresholdDontNotify, thresholdDuration, thresholdEnabled, thresholdInterval, thresholdLimit),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(threshold))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(threshold),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "update"}, scenarios)
}

var listThresholdString = strings.TrimSpace(`
Signal         Name              ID            Action  Enabled  Do Not Notify  Limit  Interval  Duration  Created At
test-signal    Test_Threshold    thresholdID   block   true     false          10     3600      86400     2021-06-15T23:00:00Z
test-signal-2  Test_Threshold_2  thresholdID2  log     false    true           20     600       43200     2021-06-15T23:00:00Z
`) + "\n"

var zeroListThresholdString = strings.TrimSpace(`
Signal  Name  ID  Action  Enabled  Do Not Notify  Limit  Interval  Duration  Created At
`) + "\n"

var thresholdString = strings.TrimSpace(`
Signal: test-signal
Name: Test_Threshold
Action: block
Do Not Notify: false
Duration: 86400
Enabled: true
Interval: 3600
Limit: 10
`)
