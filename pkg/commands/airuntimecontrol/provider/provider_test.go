package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/airuntimecontrol"
	sub "github.com/fastly/cli/pkg/commands/airuntimecontrol/provider"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/provider"
)

func TestProviderList(t *testing.T) {
	providers := provider.Providers{
		Data: []provider.Provider{
			{
				ID:             "anthropic",
				DisplayName:    "Anthropic",
				DefaultBaseURL: "https://api.anthropic.com",
				Models: []provider.Model{
					{ID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4", ProviderID: "anthropic"},
				},
			},
		},
		Meta: provider.Meta{Total: 1},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(providers))),
					},
				},
			},
			WantOutputs: []string{"anthropic", "Anthropic"},
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(providers))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(providers),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestProviderListModels(t *testing.T) {
	models := provider.Models{
		Data: []provider.Model{
			{ID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4", ProviderID: "anthropic"},
		},
		Meta: provider.Meta{Total: 1},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --provider-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --provider-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--provider-id %s", "anthropic"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(models))),
					},
				},
			},
			WantOutputs: []string{"claude-sonnet-4-20250514", "Claude Sonnet 4"},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-models"}, scenarios)
}
