package bot_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/bot"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

var bot = policy.Bot{
	BotID:  "googlebot",
	Name:   "Googlebot",
	Action: policy.Action{Type: "inherit"},
}

func TestBotList(t *testing.T) {
	// Listing across the workspace returns each bot's category.
	workspaceWide := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(policy.Bots{
		Data: []policy.Bot{
			{BotID: "googlebot", Name: "Googlebot", CategoryID: "search-engines", Action: policy.Action{Type: "inherit"}},
			{BotID: "scrapy", Name: "Scrapy", CategoryID: "scrapers", Action: policy.Action{Type: "block"}},
		},
	})}
	// Listing within a category omits it.
	categoryBots := policy.Bots{Data: []policy.Bot{bot}}
	inCategory := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(categoryBots)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:   "lists all bots in the workspace",
			Args:   "--workspace-id ws123",
			Client: workspaceWide.Client(),
			WantOutput: strings.TrimSpace(`
ID         Name       Category ID     Action
googlebot  Googlebot  search-engines  inherit
scrapy     Scrapy     scrapers        block
`) + "\n",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if workspaceWide.Path != "/bot-management/v1/workspaces/ws123/policy/bots" {
					t.Errorf("unexpected path: %s", workspaceWide.Path)
				}
			},
		},
		{
			Name:   "--category-id lists the category's bots and fills in the category column",
			Args:   "--workspace-id ws123 --category-id search-engines",
			Client: inCategory.Client(),
			WantOutput: strings.TrimSpace(`
ID         Name       Category ID     Action
googlebot  Googlebot  search-engines  inherit
`) + "\n",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if inCategory.Path != "/bot-management/v1/workspaces/ws123/policy/categories/search-engines/bots" {
					t.Errorf("unexpected path: %s", inCategory.Path)
				}
			},
		},
		{
			Name:       "--json emits the bots exactly as returned",
			Args:       "--workspace-id ws123 --category-id search-engines --json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(categoryBots)}).Client(),
			WantOutput: fstfmt.EncodeJSON(categoryBots.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestBotDescribe(t *testing.T) {
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(bot)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--bot-id is required",
			Args:      "--workspace-id ws123 --category-id search-engines",
			WantError: "error parsing arguments: required flag --bot-id not provided",
		},
		{
			Name:   "describes the bot",
			Args:   "--workspace-id ws123 --category-id search-engines --bot-id googlebot",
			Client: rec.Client(),
			WantOutput: strings.TrimSpace(`
ID: googlebot
Name: Googlebot
Action: inherit
`),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodGet || rec.Path != "/bot-management/v1/workspaces/ws123/policy/categories/search-engines/bots/googlebot" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestBotUpdate(t *testing.T) {
	updated := bot
	updated.Action.Type = "block"
	rec := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(updated)}
	inherit := &testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(bot)}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--action rejects unknown values",
			Args:      "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action deny",
			WantError: "enum value must be one of allow,block,challenge,inherit, got 'deny'",
		},
		{
			Name:       "sets the bot action",
			Args:       "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action block",
			Client:     rec.Client(),
			WantOutput: fstfmt.Success("Updated bot 'Googlebot' (bot-id: googlebot, action: block)"),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				if rec.Method != http.MethodPatch || rec.Path != "/bot-management/v1/workspaces/ws123/policy/categories/search-engines/bots/googlebot" {
					t.Errorf("unexpected request: %s %s", rec.Method, rec.Path)
				}
				testutil.AssertEqual(t, `{"action":{"type":"block"}}`, strings.TrimSpace(string(rec.RequestBody)))
			},
		},
		{
			Name:       "bots can inherit their category's action",
			Args:       "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action inherit",
			Client:     inherit.Client(),
			WantOutput: fstfmt.Success("Updated bot 'Googlebot' (bot-id: googlebot, action: inherit)"),
			Validator: func(t *testing.T, _ *testutil.CLIScenario, _ *global.Data, _ *threadsafe.Buffer) {
				testutil.AssertEqual(t, `{"action":{"type":"inherit"}}`, strings.TrimSpace(string(inherit.RequestBody)))
			},
		},
		{
			Name:       "--json emits the updated bot",
			Args:       "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action block --json",
			Client:     (&testutil.RequestRecorder{Status: http.StatusOK, Body: testutil.GenJSON(updated)}).Client(),
			WantOutput: fstfmt.EncodeJSON(updated),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
