package tsigkey_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/dns/v1/tsigkeys"

	dnsroot "github.com/fastly/cli/pkg/commands/dns"
	root "github.com/fastly/cli/pkg/commands/dns/tsigkey"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

const (
	tsigKeyID   = "tsig-key-id-123"
	tsigKeyName = "my-tsig-key"
	tsigAlgo    = "hmac-sha256"
	tsigSecret  = "dGVzdHNlY3JldA==" // #nosec G101 (CWE-798)
)

var testKey = tsigkeys.TSIGKey{
	ID:        fastly.ToPointer(tsigKeyID),
	Name:      fastly.ToPointer(tsigKeyName),
	Algorithm: fastly.ToPointer(tsigAlgo),
}

func TestTSIGKeyCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag",
		},
		{
			Args: fmt.Sprintf("--name %s --algorithm %s --secret %s", tsigKeyName, tsigAlgo, tsigSecret),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testKey))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created TSIG key '%s' (tsig-key-id: %s)", tsigKeyName, tsigKeyID),
		},
		{
			Args:      fmt.Sprintf("--name `invalid name` --algorithm %s --secret %s", tsigAlgo, tsigSecret),
			WantError: "TSIG key names cannot contain spaces",
		},
		{
			Args:      fmt.Sprintf("--name %s --algorithm %s --secret %s", strings.Repeat("a", 256), tsigAlgo, tsigSecret),
			WantError: "TSIG key names cannot exceed 255 characters",
		},
		{
			Args:      fmt.Sprintf("--verbose --json --name %s --algorithm %s --secret %s", tsigKeyName, tsigAlgo, tsigSecret),
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: fmt.Sprintf("--name %s --algorithm %s --secret %s --json", tsigKeyName, tsigAlgo, tsigSecret),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testKey))),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(testKey)),
		},
		{
			Args: fmt.Sprintf("--name %s --algorithm %s --secret %s", tsigKeyName, tsigAlgo, tsigSecret),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Bad Request"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "create"}, scenarios)
}

func TestTSIGKeyDescribe(t *testing.T) {
	resp := testutil.GenJSON(testKey)

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --tsig-key-id not provided",
		},
		{
			Args:      "--verbose --json --tsig-key-id " + tsigKeyID,
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "--tsig-key-id " + tsigKeyID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: fmtKey(&testKey),
		},
		{
			Args: "--tsig-key-id " + tsigKeyID + " --json",
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
			Args: "--tsig-key-id " + tsigKeyID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Not Found"}]}`)),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "describe"}, scenarios)
}

func TestTSIGKeyDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --tsig-key-id not provided",
		},
		{
			Args: "--tsig-key-id " + tsigKeyID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       http.NoBody,
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deleted TSIG key (tsig-key-id: %s)", tsigKeyID),
		},
		{
			Args: "--tsig-key-id " + tsigKeyID + " --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       http.NoBody,
					},
				},
			},
			WantOutput: fmt.Sprintf(`"id": %q`, tsigKeyID),
		},
		{
			Args: "--tsig-key-id " + tsigKeyID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Not Found"}]}`)),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "delete"}, scenarios)
}

func TestTSIGKeyList(t *testing.T) {
	keys := []tsigkeys.TSIGKey{testKey}
	resp := testutil.GenJSON(tsigkeys.TSIGKeys{
		Data: keys,
		Meta: tsigkeys.MetaTSIGKeys{},
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: tsigKeyName,
		},
		{
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(keys)),
		},
		{
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Bad Request"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "list"}, scenarios)
}

func TestTSIGKeyUpdate(t *testing.T) {
	updated := tsigkeys.TSIGKey{
		ID:        fastly.ToPointer(tsigKeyID),
		Name:      fastly.ToPointer("new-name"),
		Algorithm: fastly.ToPointer(tsigAlgo),
	}

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --tsig-key-id not provided",
		},
		{
			Args:      "--verbose --json --tsig-key-id " + tsigKeyID,
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args:      fmt.Sprintf("--tsig-key-id %s --name `invalid name`", tsigKeyID),
			WantError: "TSIG key names cannot contain spaces",
		},
		{
			Args:      fmt.Sprintf("--tsig-key-id %s --name %s", tsigKeyID, strings.Repeat("a", 256)),
			WantError: "TSIG key names cannot exceed 255 characters",
		},
		{
			Args: "--tsig-key-id " + tsigKeyID + " --name new-name",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated TSIG key '%s' (tsig-key-id: %s)", "new-name", tsigKeyID),
		},
		{
			Args: "--tsig-key-id " + tsigKeyID + " --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(updated)),
		},
		{
			Args: "--tsig-key-id " + tsigKeyID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Bad Request"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "update"}, scenarios)
}

func fmtKey(k *tsigkeys.TSIGKey) string {
	var b bytes.Buffer
	text.PrintTSIGKey(&b, "", k)
	return b.String()
}
