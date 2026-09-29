package category_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/category"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

var category = policy.Category{
	CategoryID:  "search-engines",
	Name:        "Search Engines",
	Description: "Crawlers operated by search providers",
	Action:      policy.Action{Type: "allow"},
}

func TestCategoryList(t *testing.T) {
	page := policy.Categories{
		Data: []policy.Category{
			category,
			{CategoryID: "scrapers", Name: "Scrapers", Action: policy.Action{Type: "block"}},
		},
	}
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(page)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:   "lists categories as a table",
			Args:   "--workspace-id ws123",
			Client: rec.Client(),
			WantOutput: strings.TrimSpace(`
ID              Name            Action
search-engines  Search Engines  allow
scrapers        Scrapers        block
`) + "\n",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Path != "/bot-management/v1/workspaces/ws123/policy/categories" {
					t.Errorf("unexpected path: %s", rec.Path)
				}
			},
		},
		{
			Name:       "--json emits the categories",
			Args:       "--workspace-id ws123 --json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(page)}).Client(),
			WantOutput: fstfmt.EncodeJSON(page.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestCategoryDescribe(t *testing.T) {
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(category)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--category-id is required",
			Args:      "--workspace-id ws123",
			WantError: "error parsing arguments: required flag --category-id not provided",
		},
		{
			Name:   "describes the category",
			Args:   "--workspace-id ws123 --category-id search-engines",
			Client: rec.Client(),
			WantOutput: strings.TrimSpace(`
ID: search-engines
Name: Search Engines
Description: Crawlers operated by search providers
Action: allow
`),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodGet || rec.Path != "/bot-management/v1/workspaces/ws123/policy/categories/search-engines" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestCategoryUpdate(t *testing.T) {
	updated := category
	updated.Action.Type = "challenge"
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(updated)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--action is required",
			Args:      "--workspace-id ws123 --category-id search-engines",
			WantError: "error parsing arguments: required flag --action not provided",
		},
		{
			Name:      "categories cannot inherit",
			Args:      "--workspace-id ws123 --category-id search-engines --action inherit",
			WantError: "enum value must be one of allow,block,challenge, got 'inherit'",
		},
		{
			Name:       "sets the category action",
			Args:       "--workspace-id ws123 --category-id search-engines --action challenge",
			Client:     rec.Client(),
			WantOutput: fstfmt.Success("Updated bot category 'Search Engines' (category-id: search-engines, action: challenge)"),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodPatch || rec.Path != "/bot-management/v1/workspaces/ws123/policy/categories/search-engines" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
				testutil.AssertEqual(t, `{"action":{"type":"challenge"}}`, strings.TrimSpace(string(rec.RequestBody)))
			},
		},
		{
			Name: "API error is surfaced",
			Args: "--workspace-id ws123 --category-id nope --action block",
			Client: (&testutil.RequestRecorder{
				Status: http.StatusNotFound,
				Body:   []byte(`{"title":"not found","status":404}`),
			}).Client(),
			WantError: "404 - Not Found",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
