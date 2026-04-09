package tags_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/apisecurity"
	sub "github.com/fastly/cli/pkg/commands/apisecurity/tags"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v14/fastly/apisecurity/operations"
)

const (
	serviceID      = "test-service-id"
	tagID          = "tag-123"
	tagName        = "APIv1"
	tagDescription = "All-APIv1-endpoints"
	updatedTagName = "APIv1.1"
	updatedTagDesc = "Updated-APIv1-endpoints"
)

var tag = operations.OperationTag{
	ID:          tagID,
	Name:        tagName,
	Description: tagDescription,
	Count:       5,
	CreatedAt:   "2021-06-15T23:00:00Z",
	UpdatedAt:   "2021-06-15T23:00:00Z",
}

func TestTagsCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      fmt.Sprintf("--name %s", tagName),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--service-id %s", serviceID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--service-id %s --name %s", serviceID, tagName),
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
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s --name %s --description %s", serviceID, tagName, tagDescription),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tag))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created operation tag '%s' (id: %s)", tagName, tagID),
		},
		{
			Name: "validate API success without description",
			Args: fmt.Sprintf("--service-id %s --name %s", serviceID, tagName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tag))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created operation tag '%s' (id: %s)", tagName, tagID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --name %s --description %s --json", serviceID, tagName, tagDescription),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tag))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(tag),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestTagsDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      fmt.Sprintf("--tag-id %s", tagID),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --tag-id flag",
			Args:      fmt.Sprintf("--service-id %s", serviceID),
			WantError: "error parsing arguments: required flag --tag-id not provided",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--service-id %s --tag-id %s", serviceID, tagID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid tag ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s --tag-id %s", serviceID, tagID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted operation tag (id: %s)", tagID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --tag-id %s --json", serviceID, tagID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"service_id": %q, "tag_id": %q, "deleted": true}`, serviceID, tagID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestTagsGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      fmt.Sprintf("--tag-id %s", tagID),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --tag-id flag",
			Args:      fmt.Sprintf("--service-id %s", serviceID),
			WantError: "error parsing arguments: required flag --tag-id not provided",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--service-id %s --tag-id invalid", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid tag ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s --tag-id %s", serviceID, tagID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tag))),
					},
				},
			},
			WantOutput: tagString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --tag-id %s --json", serviceID, tagID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tag))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(tag),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestTagsList(t *testing.T) {
	tagsObject := operations.OperationTags{
		Data: []operations.OperationTag{
			{
				ID:          "tag-001",
				Name:        "API v1",
				Description: "All v1 endpoints",
				Count:       10,
				CreatedAt:   "2021-06-15T23:00:00Z",
				UpdatedAt:   "2021-06-15T23:00:00Z",
			},
			{
				ID:          "tag-002",
				Name:        "API v2",
				Description: "All v2 endpoints",
				Count:       25,
				CreatedAt:   "2021-07-01T12:00:00Z",
				UpdatedAt:   "2021-07-01T12:00:00Z",
			},
		},
		Meta: operations.Meta{
			Limit: 50,
			Total: 2,
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--service-id %s", serviceID),
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
			Name: "validate API success (zero tags)",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(operations.OperationTags{
							Data: []operations.OperationTag{},
							Meta: operations.Meta{
								Limit: 50,
								Total: 0,
							},
						}))),
					},
				},
			},
			WantOutput: zeroListTagsString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tagsObject))),
					},
				},
			},
			WantOutput: listTagsString,
		},
		{
			Name: "validate API success with pagination",
			Args: fmt.Sprintf("--service-id %s", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tagsObject))),
					},
				},
			},
			WantOutput: listTagsString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --json", serviceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(tagsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(tagsObject.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestTagsUpdate(t *testing.T) {
	updatedTag := operations.OperationTag{
		ID:          tagID,
		Name:        updatedTagName,
		Description: updatedTagDesc,
		Count:       5,
		CreatedAt:   "2021-06-15T23:00:00Z",
		UpdatedAt:   "2021-06-16T10:00:00Z",
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      fmt.Sprintf("--tag-id %s --name %s --description %s", tagID, updatedTagName, updatedTagDesc),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --tag-id flag",
			Args:      fmt.Sprintf("--service-id %s --name %s --description %s", serviceID, updatedTagName, updatedTagDesc),
			WantError: "error parsing arguments: required flag --tag-id not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--service-id %s --tag-id %s --description %s", serviceID, tagID, updatedTagDesc),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --description flag",
			Args:      fmt.Sprintf("--service-id %s --tag-id %s --name %s", serviceID, tagID, updatedTagName),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name: "validate bad request",
			Args: fmt.Sprintf("--service-id %s --tag-id %s --name %s --description %s", serviceID, tagID, updatedTagName, updatedTagDesc),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid tag",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s --tag-id %s --name %s --description %s", serviceID, tagID, updatedTagName, updatedTagDesc),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedTag))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated operation tag '%s' (id: %s)", updatedTagName, tagID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --tag-id %s --name %s --description %s --json", serviceID, tagID, updatedTagName, updatedTagDesc),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatedTag))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatedTag),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

var tagString = strings.TrimSpace(`
ID: tag-123
Name: APIv1
Description: All-APIv1-endpoints
Operation Count: 5
Created At: 2021-06-15T23:00:00Z
Updated At: 2021-06-15T23:00:00Z
`) + "\n"

var listTagsString = strings.TrimSpace(`
ID       Name    Description       Operations  Created At            Updated At
tag-001  API v1  All v1 endpoints  10          2021-06-15T23:00:00Z  2021-06-15T23:00:00Z
tag-002  API v2  All v2 endpoints  25          2021-07-01T12:00:00Z  2021-07-01T12:00:00Z
`) + "\n"

var zeroListTagsString = strings.TrimSpace(`
ID  Name  Description  Operations  Created At  Updated At
`) + "\n"
