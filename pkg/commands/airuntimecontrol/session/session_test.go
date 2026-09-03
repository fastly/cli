package session_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/airuntimecontrol"
	sub "github.com/fastly/cli/pkg/commands/airuntimecontrol/session"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/session"
)

func TestSessionList(t *testing.T) {
	sessions := session.Sessions{
		Data: []session.Session{
			{
				ID:             "5f8a3b2c1d0e9f8a7b6c5d4e",
				VirtualKeyID:   "75352aad10d9828b8de7",
				VirtualKeyName: "go-fastly-test-key",
				Model:          "claude-sonnet-4-20250514",
				Provider:       "Anthropic",
				Requests:       3,
				InputTokens:    120,
				OutputTokens:   240,
				CreatedAt:      "2026-07-29T16:22:36Z",
				UpdatedAt:      "2026-07-29T16:22:36Z",
			},
		},
		Meta: session.Meta{Total: 1},
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
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(sessions))),
					},
				},
			},
			WantOutputs: []string{sessions.Data[0].ID, sessions.Data[0].VirtualKeyName},
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(sessions))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(sessions.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}
