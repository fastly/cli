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
	AuthType:  "api-key",
	CreatedAt: "2026-07-29T16:22:36Z",
	UpdatedAt: "2026-07-29T16:22:36Z",
}

func TestProviderConnectionCreate(t *testing.T) {
	bedrockConnection := providerconnection.ProviderConnection{
		ID:       connID,
		Name:     "bedrock",
		Models:   []string{"claude"},
		AuthType: "aws-iam",
		Region:   "us-east-1",
	}
	vertexConnection := providerconnection.ProviderConnection{
		ID:       connID,
		Name:     "vertex",
		Models:   []string{"gemini"},
		AuthType: "gcp-iam",
		Region:   "us-central1",
	}

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
			Name:      "validate missing --api-key flag",
			Args:      fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s", name, baseURL),
			WantError: "error parsing arguments: required flag --api-key not provided for --auth-type api-key",
		},
		{
			Name:      "validate missing --base-url flag",
			Args:      fmt.Sprintf("--name %s --models gpt-4,gpt-4o --api-key %s", name, apiKey),
			WantError: "error parsing arguments: required flag --base-url not provided for --auth-type api-key",
		},
		{
			Name:      "validate --pool-id is not supported for api-key",
			Args:      fmt.Sprintf("--name %s --models gpt-4,gpt-4o --base-url %s --api-key %s --pool-id example-pool", name, baseURL, apiKey),
			WantError: "error parsing arguments: --pool-id is not supported with --auth-type api-key",
		},
		{
			Name:      "validate --api-key is not supported for aws-iam",
			Args:      fmt.Sprintf("--name bedrock --models claude --auth-type aws-iam --customer-role-arn arn:aws:iam::123456789012:role/example --region us-east-1 --api-key %s", apiKey),
			WantError: "error parsing arguments: --api-key is not supported with --auth-type aws-iam",
		},
		{
			Name:      "validate missing --region flag for aws-iam",
			Args:      "--name bedrock --models claude --auth-type aws-iam --customer-role-arn arn:aws:iam::123456789012:role/example",
			WantError: "error parsing arguments: required flag --region not provided for --auth-type aws-iam",
		},
		{
			Name: "validate aws-iam API success",
			Args: "--name bedrock --models claude --auth-type aws-iam --customer-role-arn arn:aws:iam::123456789012:role/example --region us-east-1",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(bedrockConnection))),
					},
				},
			},
			WantOutputs:    []string{"Auth Type: aws-iam", "Region: us-east-1"},
			DontWantOutput: "Base URL:",
		},
		{
			Name:      "validate missing --project-number flag for gcp-iam",
			Args:      "--name vertex --models gemini --auth-type gcp-iam --pool-id example-pool --workload-provider-id example-provider --service-account-email example@example.iam.gserviceaccount.com --project-id my-project --region us-central1",
			WantError: "error parsing arguments: required flag --project-number not provided for --auth-type gcp-iam",
		},
		{
			Name:      "validate --customer-role-arn is not supported for gcp-iam",
			Args:      "--name vertex --models gemini --auth-type gcp-iam --pool-id example-pool --workload-provider-id example-provider --service-account-email example@example.iam.gserviceaccount.com --project-id my-project --project-number 123456789012 --region us-central1 --customer-role-arn arn:aws:iam::123456789012:role/example",
			WantError: "error parsing arguments: --customer-role-arn is not supported with --auth-type gcp-iam",
		},
		{
			Name: "validate gcp-iam API success",
			Args: "--name vertex --models gemini --auth-type gcp-iam --pool-id example-pool --workload-provider-id example-provider --service-account-email example@example.iam.gserviceaccount.com --project-id my-project --project-number 123456789012 --region us-central1",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(vertexConnection))),
					},
				},
			},
			WantOutputs:    []string{"Name: vertex", "Auth Type: gcp-iam", "Region: us-central1"},
			DontWantOutput: "Base URL:",
		},
		{
			Name:      "validate --name for aws-iam",
			Args:      "--name my-bedrock --models claude --auth-type aws-iam --customer-role-arn arn:aws:iam::123456789012:role/example --region us-east-1",
			WantError: "error parsing arguments: --name must be bedrock for --auth-type aws-iam",
		},
		{
			Name:      "validate --name for gcp-iam",
			Args:      "--name my-vertex --models gemini --auth-type gcp-iam",
			WantError: "error parsing arguments: --name must be vertex for --auth-type gcp-iam",
		},
		{
			Name:      "validate invalid --auth-type value",
			Args:      fmt.Sprintf("--name %s --models gpt-4 --auth-type oauth", name),
			WantError: "error parsing arguments: enum value must be one of api-key,aws-iam,gcp-iam, got 'oauth'",
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
