package providerconnection_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/airuntimecontrol"
	sub "github.com/fastly/cli/pkg/commands/airuntimecontrol/providerconnection"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

const (
	connID  = "3n9xK2mQpL8vRtY4wBzC7d"
	name    = "my-openai-connection"
	baseURL = "https://api.openai.com/v1"
	apiKey  = "sk-test-1234567890"
)

var connection = providerconnection.ProviderConnection{
	ID:        connID,
	Name:      name,
	Models:    []string{"gpt-4", "gpt-4o"},
	BaseURL:   baseURL,
	CreatedAt: "2026-07-29T16:22:36Z",
	UpdatedAt: "2026-07-29T16:22:36Z",
}

func TestProviderConnectionCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--models gpt-4 --base-url %s --api-key %s", baseURL, apiKey),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --models flag",
			Args:      fmt.Sprintf("--name %s --base-url %s --api-key %s", name, baseURL, apiKey),
			WantError: "error parsing arguments: required flag --models not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s --api-key %s", name, baseURL, apiKey),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connection))),
					},
				},
			},
			WantOutputs: []string{"ID: " + connID, "Name: " + name},
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s --api-key %s --json", name, baseURL, apiKey),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connection))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(connection),
		},
		{
			Name:      "validate missing --api-key flag and no FASTLY_ARC_API_KEY env var",
			Args:      fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s", name, baseURL),
			WantError: "error reading provider API key: no API key found",
		},
		{
			Name:    "validate --api-key falls back to FASTLY_ARC_API_KEY env var",
			Args:    fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s", name, baseURL),
			EnvVars: map[string]string{"FASTLY_ARC_API_KEY": apiKey},
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connection))),
					},
				},
			},
			WantOutputs: []string{"ID: " + connID, "Name: " + name},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestProviderConnectionGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--id %s", connID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connection))),
					},
				},
			},
			WantOutput: "ID: " + connID,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestProviderConnectionList(t *testing.T) {
	connections := providerconnection.ProviderConnections{
		Data: []providerconnection.ProviderConnection{connection},
		Meta: providerconnection.Meta{Total: 1},
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
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connections))),
					},
				},
			},
			WantOutputs: []string{connID, name},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestProviderConnectionUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "--base-url https://example.com",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--id %s --base-url %s", connID, baseURL),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(connection))),
					},
				},
			},
			WantOutput: "Base URL: " + baseURL,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestProviderConnectionDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--id %s", connID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       io.NopCloser(bytes.NewReader(nil)),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted provider connection (id: %s)", connID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}
