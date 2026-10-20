package key_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/airuntimecontrol"
	sub "github.com/fastly/cli/pkg/commands/airuntimecontrol/key"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
)

const (
	keyID       = "75352aad10d9828b8de7"
	virtualKey  = "go-fastly-test-key"
	model       = "claude-sonnet-4-20250514"
	provider    = "Anthropic"
	userID      = "6YHKQR3CxW7dGb67TNV4IB"
	accessToken = "arc_sk_go-fastly-test" // #nosec G101
)

var expiresAtStr = "2026-07-30T16:22:36Z"

var virtualKeyWithToken = key.VirtualKeyWithToken{
	VirtualKey: key.VirtualKey{
		ID:         keyID,
		Name:       virtualKey,
		UserID:     userID,
		CustomerID: "4dOhK7xEkULiabovMyhRBu",
		Model:      model,
		Provider:   provider,
		ExpiresAt:  &expiresAtStr,
		CreatedAt:  "2026-07-29T16:22:36Z",
		UpdatedAt:  "2026-07-29T16:22:36Z",
	},
	AccessToken: accessToken,
}

func TestKeyCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--model %s --provider %s", model, provider),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate API success without --model",
			Args: fmt.Sprintf("--name %s --provider %s", virtualKey, provider),
			Client: &http.Client{
				Transport: &captureRoundTripper{
					t:    t,
					want: fmt.Sprintf(`{"name":%q,"provider":%q}`, virtualKey, provider),
					body: testutil.GenJSON(virtualKeyWithToken),
				},
			},
			WantOutput: "Access Token: " + accessToken,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--name %s --model %s --provider %s", virtualKey, model, provider),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualKeyWithToken))),
					},
				},
			},
			WantOutputs: []string{"Access Token: " + accessToken, "ID: " + keyID},
		},
		{
			Name: "validate AI Firewall and rate limit flags",
			Args: fmt.Sprintf("--name %s --model %s --provider %s --security-enabled true --security-action block --rpm-limit 60 --tpm-limit 100000", virtualKey, model, provider),
			Client: &http.Client{
				Transport: &captureRoundTripper{
					t:    t,
					want: fmt.Sprintf(`"model":%q,"provider":%q,"security_enabled":true,"security_action":"block","rpm_limit":60,"tpm_limit":100000`, model, provider),
					body: testutil.GenJSON(virtualKeyWithToken),
				},
			},
			WantOutput: "Access Token: " + accessToken,
		},
		{
			Name:      "validate invalid --security-action value",
			Args:      fmt.Sprintf("--name %s --model %s --provider %s --security-action deny", virtualKey, model, provider),
			WantError: "error parsing arguments: enum value must be one of log,block,replace, got 'deny'",
		},
		{
			Name:      "validate invalid --security-enabled value",
			Args:      fmt.Sprintf("--name %s --model %s --provider %s --security-enabled yes", virtualKey, model, provider),
			WantError: "'security-enabled' flag must be one of the following [true, false]",
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --model %s --provider %s --json", virtualKey, model, provider),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualKeyWithToken))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(virtualKeyWithToken),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestKeyGet(t *testing.T) {
	listItem := key.VirtualKeyListItem{
		ID:       keyID,
		Name:     virtualKey,
		Model:    model,
		Provider: provider,
		UserID:   userID,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s", keyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listItem))),
					},
				},
			},
			WantOutput: fmt.Sprintf("ID: %s", keyID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestKeyList(t *testing.T) {
	keys := key.VirtualKeys{
		Data: []key.VirtualKeyListItem{
			{ID: keyID, Name: virtualKey, Model: model, Provider: provider},
		},
		Meta: key.Meta{Total: 1},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate API success",
			Args: "--non-interactive",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(keys))),
					},
				},
			},
			WantOutputs: []string{keyID, virtualKey},
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(keys))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(keys.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestKeyUpdate(t *testing.T) {
	tpmLimit := 5000
	updated := key.VirtualKey{
		ID:                keyID,
		Name:              "go-fastly-test-key-updated",
		UserID:            userID,
		Model:             model,
		Provider:          provider,
		TPMLimit:          &tpmLimit,
		SpendLimitEnabled: true,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      "--name updated",
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s --name %s", keyID, updated.Name),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fmt.Sprintf("Name: %s", updated.Name),
		},
		{
			Name: "validate AI Firewall, rate limit, and spend limit flags",
			Args: fmt.Sprintf("--key-id %s --security-enabled false --security-action log --rpm-limit 0 --tpm-limit 5000 --spend-limit-enabled true", keyID),
			Client: &http.Client{
				Transport: &captureRoundTripper{
					t:    t,
					want: `"security_enabled":false,"security_action":"log","rpm_limit":null,"tpm_limit":5000,"spend_limit_enabled":true`,
					body: testutil.GenJSON(updated),
				},
			},
			WantOutputs: []string{"TPM Limit: 5000", "Spend Limit Enabled: true"},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestKeyDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s", keyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       io.NopCloser(bytes.NewReader(nil)),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted virtual key (key-id: %s)", keyID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestKeyRotate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      fmt.Sprintf("--expires-at %s", expiresAtStr),
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name: "validate API success without --expires-at",
			Args: fmt.Sprintf("--key-id %s", keyID),
			Client: &http.Client{
				Transport: &captureRoundTripper{
					t:    t,
					want: `{}`,
					body: testutil.GenJSON(virtualKeyWithToken),
				},
			},
			WantOutput: "Access Token: " + accessToken,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s --expires-at %s", keyID, expiresAtStr),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualKeyWithToken))),
					},
				},
			},
			WantOutput: "Access Token: " + accessToken,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "rotate"}, scenarios)
}

func TestKeyGetFailoverPolicy(t *testing.T) {
	policy := key.FailoverPolicy{
		Targets: []key.FailoverTarget{
			{ProviderConnectionID: "pTfnUUBcU1yv8SoRJn5PZ2"},
			{ProviderConnectionID: "3n9xK2mQpL8vRtY4wBzC7d"},
		},
		TimeoutEnabled:      true,
		RateLimit429Enabled: true,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s", keyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(policy))),
					},
				},
			},
			WantOutputs: []string{
				"Targets: pTfnUUBcU1yv8SoRJn5PZ2, 3n9xK2mQpL8vRtY4wBzC7d",
				"Timeout Enabled: true",
				"Rate Limit 429 Enabled: true",
				"Server Error 5xx Enabled: false",
			},
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--key-id %s --json", keyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(policy))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(policy),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get-failover-policy"}, scenarios)
}

func TestKeyUpdateFailoverPolicy(t *testing.T) {
	policy := key.FailoverPolicy{
		Targets: []key.FailoverTarget{
			{ProviderConnectionID: "pTfnUUBcU1yv8SoRJn5PZ2"},
			{ProviderConnectionID: "3n9xK2mQpL8vRtY4wBzC7d"},
		},
		ServerError5xxEnabled: true,
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key-id flag",
			Args:      "--targets pTfnUUBcU1yv8SoRJn5PZ2",
			WantError: "error parsing arguments: required flag --key-id not provided",
		},
		{
			Name:      "validate missing --targets flag",
			Args:      fmt.Sprintf("--key-id %s", keyID),
			WantError: "error parsing arguments: required flag --targets not provided",
		},
		{
			Name:      "validate invalid --timeout-enabled value",
			Args:      fmt.Sprintf("--key-id %s --targets pTfnUUBcU1yv8SoRJn5PZ2 --timeout-enabled maybe", keyID),
			WantError: "'timeout-enabled' flag must be one of the following [true, false]",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--key-id %s --targets pTfnUUBcU1yv8SoRJn5PZ2,3n9xK2mQpL8vRtY4wBzC7d --timeout-enabled false --rate-limit-429-enabled false --server-error-5xx-enabled true", keyID),
			Client: &http.Client{
				Transport: &captureRoundTripper{
					t:    t,
					want: `{"targets":[{"provider_connection_id":"pTfnUUBcU1yv8SoRJn5PZ2"},{"provider_connection_id":"3n9xK2mQpL8vRtY4wBzC7d"}],"timeout_enabled":false,"rate_limit_429_enabled":false,"server_error_5xx_enabled":true}`,
					body: testutil.GenJSON(policy),
				},
			},
			WantOutputs: []string{
				"Targets: pTfnUUBcU1yv8SoRJn5PZ2, 3n9xK2mQpL8vRtY4wBzC7d",
				"Server Error 5xx Enabled: true",
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update-failover-policy"}, scenarios)
}

// captureRoundTripper asserts that the request body contains want, then
// responds with body.
type captureRoundTripper struct {
	t    *testing.T
	want string
	body []byte
}

// RoundTrip implements [http.RoundTripper].
func (c *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	got, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(got), c.want) {
		c.t.Errorf("request body %s does not contain %s", got, c.want)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(bytes.NewReader(c.body)),
	}, nil
}
