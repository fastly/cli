package key_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
			Args:      fmt.Sprintf("--model %s --provider %s --user-id %s", model, provider, userID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --model flag",
			Args:      fmt.Sprintf("--name %s --provider %s --user-id %s", virtualKey, provider, userID),
			WantError: "error parsing arguments: required flag --model not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--name %s --model %s --provider %s --user-id %s", virtualKey, model, provider, userID),
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
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --model %s --provider %s --user-id %s --json", virtualKey, model, provider, userID),
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
	updated := key.VirtualKey{
		ID:       keyID,
		Name:     "go-fastly-test-key-updated",
		UserID:   userID,
		Model:    model,
		Provider: provider,
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
			Name:      "validate missing --expires-at flag",
			Args:      fmt.Sprintf("--key-id %s", keyID),
			WantError: "error parsing arguments: required flag --expires-at not provided",
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
