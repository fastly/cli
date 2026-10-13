package rule_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	routingconfig "github.com/fastly/cli/pkg/commands/routingconfig"
	path "github.com/fastly/cli/pkg/commands/routingconfig/path"
	root "github.com/fastly/cli/pkg/commands/routingconfig/path/rule"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	rulePath        = "testdata/rule.json"
	defaultRulePath = "testdata/default_rule.json"
)

func cmd(leaf string) []string {
	return []string{routingconfig.CommandName, path.CommandName, root.CommandName, leaf}
}

func TestRuleCreate(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	rid := "rule-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
			WantError: "error parsing arguments: required flag --file not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --file %s", rcid, pid, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(rules.Data{
							RuleID: rid,
							Action: rules.Action{Type: "service", Value: "SU1Z0isxPaozGVKXdv0eY"},
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created rule (rule-id: %s) in path (path-id: %s)", rid, pid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --file %s", rcid, pid, defaultRulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(rules.Data{
							RuleID:    rid,
							Action:    rules.Action{Type: "service", Value: "SU1Z0isxPaozGVKXdv0eY"},
							IsDefault: true,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created rule (rule-id: %s) in path (path-id: %s)", rid, pid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --file %s", rcid, pid, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, cmd("create"), scenarios)
}

func TestRuleDescribe(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	rid := "rule-id"

	resp := testutil.GenJSON(rules.Data{
		RuleID: rid,
		Action: rules.Action{Type: "service", Value: "SU1Z0isxPaozGVKXdv0eY"},
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
			WantError: "error parsing arguments: required flag --rule-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s --json", rcid, pid, rid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: string(resp),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s", rcid, pid, rid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, cmd("describe"), scenarios)
}

func TestRuleList(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	rid := "rule-id"

	data := []rules.Data{
		{
			RuleID: rid,
			Action: rules.Action{Type: "service", Value: "SU1Z0isxPaozGVKXdv0eY"},
		},
	}
	apiResp := testutil.GenJSON(rules.Collection{
		Data: data,
		Meta: rules.Meta{Limit: 20, Total: 1},
	})
	wantOutput := testutil.GenJSON(data)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --path-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --json", rcid, pid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(apiResp)),
					},
				},
			},
			WantOutput: string(wantOutput),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, cmd("list"), scenarios)
}

func TestRuleUpdate(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	rid := "rule-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s", rcid, pid, rid),
			WantError: "error parsing arguments: required flag --file not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s --file %s", rcid, pid, rid, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(rules.Data{
							RuleID: rid,
							Action: rules.Action{Type: "service", Value: "SU1Z0isxPaozGVKXdv0eY"},
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated rule (rule-id: %s)", rid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s --file %s", rcid, pid, rid, rulePath),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, cmd("update"), scenarios)
}

func TestRuleDelete(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	rid := "rule-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
			WantError: "error parsing arguments: required flag --rule-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s", rcid, pid, rid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deleted rule (rule-id: %s)", rid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --rule-id %s", rcid, pid, rid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, cmd("delete"), scenarios)
}
