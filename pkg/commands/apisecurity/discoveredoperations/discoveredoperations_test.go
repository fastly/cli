package discoveredoperations_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apisecurity "github.com/fastly/cli/pkg/commands/apisecurity"
	root "github.com/fastly/cli/pkg/commands/apisecurity/discoveredoperations"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
)

const (
	serviceID   = "test-service-id"
	operationID = "test-operation-id"
)

var (
	listResponse = operations.DiscoveredOperations{
		Data: []operations.DiscoveredOperation{
			{
				ID:         "test-operation-id",
				Method:     "GET",
				Domain:     "example.com",
				Path:       "/api/users",
				Status:     "DISCOVERED",
				RPS:        10.5,
				LastSeenAt: "2026-03-10T12:00:00Z",
				UpdatedAt:  "2026-03-10T12:00:00Z",
			},
			{
				ID:         "test-operation-id-2",
				Method:     "POST",
				Domain:     "example.com",
				Path:       "/api/users",
				Status:     "SAVED",
				RPS:        5.2,
				LastSeenAt: "2026-03-10T12:00:00Z",
				UpdatedAt:  "2026-03-10T12:00:00Z",
			},
		},
		Meta: operations.Meta{
			Limit: 2,
			Total: 2,
		},
	}

	updateResponse = operations.DiscoveredOperation{
		ID:         "test-operation-id",
		Method:     "GET",
		Domain:     "example.com",
		Path:       "/api/users",
		Status:     "IGNORED",
		RPS:        10.5,
		LastSeenAt: "2026-03-10T12:00:00Z",
		UpdatedAt:  "2026-03-10T13:00:00Z",
	}

	updateResponseJSON = testutil.GenJSON(updateResponse)
	listResponseJSON   = testutil.GenJSON(listResponse)

	bulkResponse = operations.BulkOperationResultsResponse{
		Data: []operations.BulkOperationResult{
			{
				ID:         "op-id-1",
				StatusCode: 200,
			},
			{
				ID:         "op-id-2",
				StatusCode: 200,
			},
		},
	}
	bulkResponseJSON = testutil.GenJSON(bulkResponse)

	listDiscoveredOperationsOutput = strings.TrimSpace(`
METHOD  DOMAIN       PATH        STATUS      RPS    LAST SEEN
GET     example.com  /api/users  DISCOVERED  10.50  2026-03-10T12:00:00Z
POST    example.com  /api/users  SAVED       5.20   2026-03-10T12:00:00Z
`) + "\n"

	listDiscoveredOperationsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

Service ID (via --service-id): test-service-id


Discovered Operation 1/2
	ID: test-operation-id
	Method: GET
	Domain: example.com
	Path: /api/users
	Status: DISCOVERED
	RPS: 10.50
	Last Seen: 2026-03-10T12:00:00Z
	Updated At: 2026-03-10T12:00:00Z

Discovered Operation 2/2
	ID: test-operation-id-2
	Method: POST
	Domain: example.com
	Path: /api/users
	Status: SAVED
	RPS: 5.20
	Last Seen: 2026-03-10T12:00:00Z
	Updated At: 2026-03-10T12:00:00Z
`) + "\n\n"
)

func TestListCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--status discovered",
			WantError: "error parsing arguments: required flag --service-id not provided",
		},
		{
			Name: "validate list without status filter",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsOutput,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsOutput,
		},
		{
			Name: "validate --json flag",
			Args: fmt.Sprintf("--service-id %s --status saved --json", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(listResponseJSON)),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(listResponse.Data)),
		},
		{
			Name: "validate invalid status",
			Args: fmt.Sprintf("--service-id %s --status invalid", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantError: "invalid status: invalid. Valid options: 'discovered', 'saved', 'ignored'",
		},
		{
			Name: "validate API error",
			Args: fmt.Sprintf("--service-id %s --status discovered", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
						Body:       io.NopCloser(strings.NewReader(`{"detail":"Internal Server Error"}`)),
					},
				},
			},
			WantError: "500",
		},
	}

	testutil.RunCLIScenarios(t, []string{apisecurity.CommandName, root.CommandName, "list"}, scenarios)
}

func TestListCommandWithFilters(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate --domain filter",
			Args: fmt.Sprintf("--service-id %s --status discovered --domain example.com", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsOutput,
		},
		{
			Name: "validate --method filter",
			Args: fmt.Sprintf("--service-id %s --status discovered --method GET", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsOutput,
		},
		{
			Name: "validate --path filter",
			Args: fmt.Sprintf("--service-id %s --status discovered --path /api/users", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsOutput,
		},
		{
			Name: "validate --verbose output",
			Args: fmt.Sprintf("--service-id %s --status discovered --verbose", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listDiscoveredOperationsVerboseOutput,
		},
		{
			Name: "validate empty results",
			Args: fmt.Sprintf("--service-id %s --status discovered", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(operations.DiscoveredOperations{
							Data: []operations.DiscoveredOperation{},
							Meta: operations.Meta{
								Limit: 0,
								Total: 0,
							},
						}))),
					},
				},
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{apisecurity.CommandName, root.CommandName, "list"}, scenarios)
}

func TestUpdateCommand(t *testing.T) {
	// Create temp file for bulk update test
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-ops.json")
	content := `{"operation_ids": ["op-id-1", "op-id-2"], "status": "ignored"}`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	discoveredResponse := operations.DiscoveredOperation{
		ID:         "test-operation-id",
		Method:     "GET",
		Domain:     "example.com",
		Path:       "/api/users",
		Status:     "DISCOVERED",
		RPS:        10.5,
		LastSeenAt: "2026-03-10T12:00:00Z",
		UpdatedAt:  "2026-03-10T13:00:00Z",
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      fmt.Sprintf("--operation-id %s --status ignored", operationID),
			WantError: "error parsing arguments: required flag --service-id not provided",
		},
		{
			Name:      "validate missing --operation-id and --file flags",
			Args:      fmt.Sprintf("--service-id %s --status ignored", serviceID),
			WantError: "error parsing arguments: must provide either --operation-id or --file",
		},
		{
			Name:      "validate missing --status flag",
			Args:      fmt.Sprintf("--service-id %s --operation-id %s", serviceID, operationID),
			WantError: "error parsing arguments: --status is required when using --operation-id",
		},
		{
			Name: "validate invalid status",
			Args: fmt.Sprintf("--service-id %s --operation-id %s --status invalid", serviceID, operationID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updateResponse))),
					},
				},
			},
			WantError: "invalid status: invalid. Valid options: 'discovered', 'ignored'",
		},
		{
			Name: "validate API success with status ignored",
			Args: fmt.Sprintf("--service-id %s --operation-id %s --status ignored", serviceID, operationID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updateResponse))),
					},
				},
			},
			WantOutputs: []string{
				"Updated discovered operation:",
				"ID: test-operation-id",
				"Method: GET",
				"Domain: example.com",
				"Path: /api/users",
				"Status: IGNORED",
			},
		},
		{
			Name: "validate API success with status discovered",
			Args: fmt.Sprintf("--service-id %s --operation-id %s --status discovered", serviceID, operationID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(discoveredResponse))),
					},
				},
			},
			WantOutputs: []string{
				"Updated discovered operation:",
				"Status: DISCOVERED",
			},
		},
		{
			Name: "validate --json flag",
			Args: fmt.Sprintf("--service-id %s --operation-id %s --status ignored --json", serviceID, operationID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(updateResponseJSON)),
					},
				},
			},
			WantOutput: string(updateResponseJSON),
		},
		{
			Name: "validate API error",
			Args: fmt.Sprintf("--service-id %s --operation-id %s --status ignored", serviceID, operationID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(strings.NewReader(`{"detail":"Not Found"}`)),
					},
				},
			},
			WantError: "404",
		},
		{
			Name: "validate bulk mode with --json flag",
			Args: fmt.Sprintf("--service-id %s --file %s --json", serviceID, testFile),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusMultiStatus,
						Status:     http.StatusText(http.StatusMultiStatus),
						Body:       io.NopCloser(bytes.NewReader(bulkResponseJSON)),
					},
				},
			},
			WantOutput: string(bulkResponseJSON),
		},
	}

	testutil.RunCLIScenarios(t, []string{apisecurity.CommandName, root.CommandName, "update"}, scenarios)
}

func TestUpdateCommandEdgeCases(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate cannot use both --operation-id and --file",
			Args:      fmt.Sprintf("--service-id %s --operation-id %s --file /tmp/test.json --status ignored", serviceID, operationID),
			WantError: "error parsing arguments: cannot use both --operation-id and --file",
		},
		{
			Name:      "validate cannot use --file with --status flag",
			Args:      fmt.Sprintf("--service-id %s --file /tmp/test.json --status ignored", serviceID),
			WantError: "error parsing arguments: cannot use both --file and --status",
		},
		{
			Name: "validate comma-separated operation IDs",
			Args: fmt.Sprintf("--service-id %s --operation-id op-id-1,op-id-2 --status ignored", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusMultiStatus,
						Status:     http.StatusText(http.StatusMultiStatus),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(bulkResponse))),
					},
				},
			},
			WantOutputs: []string{
				"Updated 2 discovered operation(s)",
			},
		},
		{
			Name: "validate --verbose with bulk update",
			Args: fmt.Sprintf("--service-id %s --operation-id op-id-1,op-id-2 --status ignored --verbose", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusMultiStatus,
						Status:     http.StatusText(http.StatusMultiStatus),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(bulkResponse))),
					},
				},
			},
			WantOutputs: []string{
				"Updated 2 discovered operation(s)",
				"Updating 2 operation(s) with status: IGNORED",
				"OPERATION ID",
				"STATUS CODE",
				"RESULT",
				"op-id-1",
				"200",
				"Success",
			},
		},
		{
			Name: "validate bulk update with mixed results",
			Args: fmt.Sprintf("--service-id %s --operation-id op-id-1,op-id-2,op-id-3 --status ignored", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusMultiStatus,
						Status:     http.StatusText(http.StatusMultiStatus),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(operations.BulkOperationResultsResponse{
							Data: []operations.BulkOperationResult{
								{ID: "op-id-1", StatusCode: 200},
								{ID: "op-id-2", StatusCode: 404, Reason: "Not Found"},
								{ID: "op-id-3", StatusCode: 200},
							},
						}))),
					},
				},
			},
			WantOutputs: []string{
				"Updated 2 discovered operation(s)",
				"1 operation(s) failed to update",
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{apisecurity.CommandName, root.CommandName, "update"}, scenarios)
}
