package rule_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/rule"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

const (
	ruleDescription = "Utility requests"
	ruleEnabled     = true
	ruleAction      = "allow"
	ruleID          = "someID"
	rulePath        = "testdata/test_rule.json"
	ruleType        = "request"
)

var rule = rules.Rule{
	CreatedAt:   testutil.Date,
	Description: ruleDescription,
	Enabled:     ruleEnabled,
	RuleID:      ruleID,
	Actions: []rules.Action{
		{
			Type: ruleAction,
		},
	},
	Type: ruleType,
	Scope: rules.Scope{
		Type:      string(scope.ScopeTypeAccount),
		AppliesTo: []string{"*"},
	},
}

func TestRuleCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --path flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --path not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--path %s", rulePath),
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
			Args: fmt.Sprintf("--path %s", rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(rule)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created account-level rule with ID %s", ruleID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--path %s --json", rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(rule))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(rule),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestRuleDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --rule-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --rule-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--rule-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid rule ID",
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
			Args: fmt.Sprintf("--rule-id %s", ruleID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted account-level rule with id: %s", ruleID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--rule-id %s --json", ruleID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, ruleID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestRuleGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --rule-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --rule-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--rule-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Rule ID",
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
			Args: fmt.Sprintf("--rule-id %s", ruleID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(rule)))),
					},
				},
			},
			WantOutput: ruleString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--rule-id %s --json", ruleID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(rule)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(rule),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestRuleList(t *testing.T) {
	rulesObject := rules.Rules{
		Data: []rules.Rule{
			{
				CreatedAt:   testutil.Date,
				Description: ruleDescription,
				Enabled:     ruleEnabled,
				RuleID:      ruleID,
				Actions: []rules.Action{
					{
						Type: ruleAction,
					},
				},
				Type: ruleType,
				Scope: rules.Scope{
					Type:      string(scope.ScopeTypeAccount),
					AppliesTo: []string{"*"},
				},
			},
			{
				CreatedAt:   testutil.Date,
				Description: ruleDescription + "2",
				Enabled:     ruleEnabled,
				RuleID:      ruleID + "2",
				Actions: []rules.Action{
					{
						Type: ruleAction,
					},
				},
				Type: ruleType,
				Scope: rules.Scope{
					Type:      string(scope.ScopeTypeAccount),
					AppliesTo: []string{"*"},
				},
			},
		},
		Meta: rules.MetaRules{},
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
			Name: "validate API success (zero account-level Rules)",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(rules.Rules{
							Data: []rules.Rule{},
							Meta: rules.MetaRules{},
						}))),
					},
				},
			},
			WantOutput: zeroListRulesString,
		},
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(rulesObject))),
					},
				},
			},
			WantOutput: listRulesString,
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(rulesObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(rulesObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestRuleUpdate(t *testing.T) {
	ruleObject := rules.Rule{
		CreatedAt:   testutil.Date,
		Description: ruleDescription,
		RuleID:      ruleID,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --rule-id flag",
			Args:      fmt.Sprintf("--path %s", rulePath),
			WantError: "error parsing arguments: required flag --rule-id not provided",
		},
		{
			Name:      "validate missing --path flag",
			Args:      fmt.Sprintf("--rule-id %s", ruleID),
			WantError: "error parsing arguments: required flag --path not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--rule-id %s --path %s", ruleID, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(ruleObject))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated account-level rule with id: %s", ruleID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--rule-id %s --path %s --json", ruleID, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(rule))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(rule),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

var listRulesString = strings.TrimSpace(`
ID       Action  Description        Enabled  Type     Scope    Updated At                     Created At
someID   allow   Utility requests   true     request  account  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
someID2  allow   Utility requests2  true     request  account  0001-01-01 00:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListRulesString = strings.TrimSpace(`
ID  Action  Description  Enabled  Type  Scope  Updated At  Created At
`) + "\n"

var ruleString = strings.TrimSpace(`
ID: someID
Action: allow
Description: Utility requests
Enabled: true
Type: request
Scope: account
Updated (UTC): 0001-01-01 00:00
Created (UTC): 2021-06-15 23:00
`)
