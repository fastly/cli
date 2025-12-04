package customsignal_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/customsignal"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"
)

const (
	customSignalDescription = "NGWAFCLICustomSignal"
	customSignalID          = "someID"
	customSignalName        = "CLICustomSignalName"
)

var customSignal = signals.Signal{
	CreatedAt:   testutil.Date,
	Description: customSignalDescription,
	Name:        customSignalName,
	SignalID:    customSignalID,
	Scope: signals.Scope{
		Type:      string(scope.ScopeTypeAccount),
		AppliesTo: []string{"*"},
	},
}

func TestCustomSignalCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--description %s", customSignalDescription),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--description %s --name %s", customSignalDescription, customSignalName),
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
			Args: fmt.Sprintf("--description %s --name %s", customSignalDescription, customSignalName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(customSignal)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created account-level custom signal '%s' (signal-id: %s)", customSignalName, customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--description %s --name %s --json", customSignalDescription, customSignalName),
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestCustomSignalDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --signal-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--signal-id bar",
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
			Args: fmt.Sprintf("--signal-id %s", customSignalID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted account-level custom signal (id: %s)", customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--signal-id %s --json", customSignalID),
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestCustomSignalGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --signal-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--signal-id baz",
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
			Args: fmt.Sprintf("--signal-id %s", customSignalID),
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
			Args: fmt.Sprintf("--signal-id %s --json", customSignalID),
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
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
					Type:      string(scope.ScopeTypeAccount),
					AppliesTo: []string{"*"},
				},
			},
			{
				CreatedAt:   testutil.Date,
				Description: customSignalDescription,
				Name:        customSignalName + "2",
				SignalID:    customSignalID + "2",
				Scope: signals.Scope{
					Type:      string(scope.ScopeTypeAccount),
					AppliesTo: []string{"*"},
				},
			},
		},
		Meta: signals.MetaSignals{},
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
			Name: "validate API success (zero account-level custom signals)",
			Args: "",
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
			Args: "",
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
			Args: "--json",
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
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
			Args:      fmt.Sprintf("--description %s", customSignalDescription),
			WantError: "error parsing arguments: required flag --signal-id not provided",
		},
		{
			Name:      "validate missing --description flag",
			Args:      fmt.Sprintf("--signal-id %s", customSignalID),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--signal-id %s --description %s", customSignalID, customSignalDescription),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(customSignalObject))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated account-level signal '%s' (signal-id: %s)", customSignalName, customSignalID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--signal-id %s --description %s --json", customSignalID, customSignalDescription),
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

var listCustomSignalsString = strings.TrimSpace(`
ID       Name                  Description           Scope    Updated At                     Created At
someID   CLICustomSignalName   NGWAFCLICustomSignal  account  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
someID2  CLICustomSignalName2  NGWAFCLICustomSignal  account  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListCustomSignalsString = strings.TrimSpace(`
ID  Name  Description  Scope  Updated At  Created At
`) + "\n"

var customSignalString = strings.TrimSpace(`
ID: someID
Name: CLICustomSignalName
Description: NGWAFCLICustomSignal
Scope: account
Updated (UTC): 0001-01-01 00:00
Created (UTC): 2021-06-15 23:00
`)
