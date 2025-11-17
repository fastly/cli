package virtualpatch_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/virtualpatch"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/virtualpatches"
)

const (
	virtualpatchID          = "CVE-2017-5638"
	virtualpatchDescription = "Apache Struts multipart/form remote execution"
	virtualpatchEnabled     = false
	virtualpatchMode        = "log"
	workspaceID             = "nBw2ENWfOY1M2dpSwK1l5R"
)

var virtualpatch = virtualpatches.VirtualPatch{
	ID:          virtualpatchID,
	Description: virtualpatchDescription,
	Enabled:     virtualpatchEnabled,
	Mode:        virtualpatchMode,
}

func TestVirtualPatchRetrieve(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--virtualpatch-id %s", virtualpatchID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --virtualpatch-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --virtualpatch-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id invalid", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "This resource does not exist",
    							"status": 404
							}
						`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(virtualpatch)))),
					},
				},
			},
			WantOutput: virtualpatchString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --json", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(virtualpatch)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(virtualpatch),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "retrieve"}, scenarios)
}

func TestVirtualPatchList(t *testing.T) {
	virtualpatchesObject := virtualpatches.VirtualPatches{
		Data: []virtualpatches.VirtualPatch{
			{
				ID:          "CVE-2024-5806",
				Description: "Progress MOVEit Transfer Authentication Bypass Vulnerability",
				Enabled:     false,
				Mode:        "log",
			},
			{
				ID:          "CVE-2024-34102",
				Description: "Adobe Commerce and Magento Open Source Unauthenticated XML Entity Injection",
				Enabled:     false,
				Mode:        "log",
			},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success (zero virtual patches)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualpatches.VirtualPatches{
							Data: []virtualpatches.VirtualPatch{},
						}))),
					},
				},
			},
			WantOutput: zeroListVirtualPatchString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualpatchesObject))),
					},
				},
			},
			WantOutput: listVirtualPatchString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(virtualpatchesObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(virtualpatchesObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestVirtualPatchUpdate(t *testing.T) {
	updatedVirtualPatch := virtualpatches.VirtualPatch{
		ID:          virtualpatchID,
		Description: virtualpatchDescription,
		Enabled:     true,
		Mode:        "block",
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--virtualpatch-id %s", virtualpatchID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name:      "validate missing --virtualpatch-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --virtualpatch-id not provided",
		},
		{
			Name: "validate not found",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id invalid --enabled true", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "This resource does not exist",
    							"status": 404
							}
						`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
		{
			Name: "validate invalid enabled flag value",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --enabled maybe", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedVirtualPatch))),
					},
				},
			},
			WantError: "'enabled' flag must be one of the following [true, false]",
		},
		{
			Name: "validate API success with enabled flag",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --enabled true", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedVirtualPatch))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated virtual patch '%s' (enabled: %t, mode: %s)", updatedVirtualPatch.ID, updatedVirtualPatch.Enabled, updatedVirtualPatch.Mode),
		},
		{
			Name: "validate API success with mode flag",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --mode block", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedVirtualPatch))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated virtual patch '%s' (enabled: %t, mode: %s)", updatedVirtualPatch.ID, updatedVirtualPatch.Enabled, updatedVirtualPatch.Mode),
		},
		{
			Name: "validate API success with both flags",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --enabled true --mode block", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedVirtualPatch))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated virtual patch '%s' (enabled: %t, mode: %s)", updatedVirtualPatch.ID, updatedVirtualPatch.Enabled, updatedVirtualPatch.Mode),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --virtualpatch-id %s --enabled true --mode block --json", workspaceID, virtualpatchID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedVirtualPatch))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatedVirtualPatch),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

var virtualpatchString = strings.TrimSpace(`
ID: CVE-2017-5638
Description: Apache Struts multipart/form remote execution
Enabled: false
Mode: log
`)

var listVirtualPatchString = strings.TrimSpace(`
ID              Description                                                                  Enabled  Mode
CVE-2024-5806   Progress MOVEit Transfer Authentication Bypass Vulnerability                 false    log
CVE-2024-34102  Adobe Commerce and Magento Open Source Unauthenticated XML Entity Injection  false    log
`) + "\n"

var zeroListVirtualPatchString = strings.TrimSpace(`
ID  Description  Enabled  Mode
`) + "\n"
