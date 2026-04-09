package operations_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v14/fastly/apisecurity/operations"

	apisecurity "github.com/fastly/cli/pkg/commands/apisecurity"
	root "github.com/fastly/cli/pkg/commands/apisecurity/operations"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	serviceID = "test-service-id"
)

var (
	listResponse = operations.Operations{
		Data: []operations.Operation{
			{
				ID:          "test-operation-id",
				Method:      "DELETE",
				Domain:      "www.foo.com",
				Path:        "/api/v1/users/{var1}",
				Description: "Retrieve user information",
				Status:      "SAVED",
				RPS:         10.5,
				CreatedAt:   "2026-02-02T14:27:16Z",
				UpdatedAt:   "2026-02-02T14:33:19Z",
				TagIDs:      []string{},
			},
			{
				ID:          "test-operation-id-2",
				Method:      "POST",
				Domain:      "www.foo.com",
				Path:        "/api/v1/users",
				Description: "Create a new user",
				Status:      "SAVED",
				RPS:         5.2,
				CreatedAt:   "2026-02-01T10:00:00Z",
				UpdatedAt:   "2026-02-01T10:30:00Z",
				TagIDs:      []string{"tag-1", "tag-2"},
			},
		},
		Meta: operations.Meta{
			Limit: 2,
			Total: 2,
		},
	}

	listResponseJSON = testutil.GenJSON(listResponse)

	listOperationsOutput = strings.TrimSpace(`
ID                   METHOD  DOMAIN       PATH                  DESCRIPTION                TAGS
test-operation-id    DELETE  www.foo.com  /api/v1/users/{var1}  Retrieve user information  0
test-operation-id-2  POST    www.foo.com  /api/v1/users         Create a new user          2
`) + "\n"

	listOperationsVerboseOutput = strings.TrimSpace(`
Operation 1/2
	ID: test-operation-id
	Method: DELETE
	Domain: www.foo.com
	Path: /api/v1/users/{var1}
	Description: Retrieve user information
	Status: SAVED
	RPS: 10.50
	Created At: 2026-02-02T14:27:16Z
	Updated At: 2026-02-02T14:33:19Z

Operation 2/2
	ID: test-operation-id-2
	Method: POST
	Domain: www.foo.com
	Path: /api/v1/users
	Description: Create a new user
	Status: SAVED
	Tag IDs: tag-1, tag-2
	RPS: 5.20
	Created At: 2026-02-01T10:00:00Z
	Updated At: 2026-02-01T10:30:00Z
`) + "\n\n"
)

func TestListCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --service-id not provided",
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
			WantOutput: listOperationsOutput,
		},
		{
			Name: "validate --json flag",
			Args: fmt.Sprintf("--service-id %s --json", serviceID),
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
			Name: "validate --verbose output",
			Args: fmt.Sprintf("--service-id %s --verbose", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listOperationsVerboseOutput,
		},
		{
			Name: "validate API error",
			Args: fmt.Sprintf("--service-id %s", serviceID),
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
			Args: fmt.Sprintf("--service-id %s --domain www.foo.com", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listOperationsOutput,
		},
		{
			Name: "validate --method filter",
			Args: fmt.Sprintf("--service-id %s --method DELETE", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listOperationsOutput,
		},
		{
			Name: "validate --path filter",
			Args: fmt.Sprintf("--service-id %s --path /api/v1/users", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listOperationsOutput,
		},
		{
			Name: "validate --tag-id filter",
			Args: fmt.Sprintf("--service-id %s --tag-id tag-1", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listResponse))),
					},
				},
			},
			WantOutput: listOperationsOutput,
		},
		{
			Name: "validate empty results",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(operations.Operations{
							Data: []operations.Operation{},
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
