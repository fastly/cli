package workspaceiplist_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspaceiplist"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

const (
	listID          = "someListID"
	listDescription = "NGWAFCLIList"
	listEntries     = "1.0.0.0"
	listType        = "ip"
	listName        = "listName"
	workspaceID     = "someWorkspaceID"
)

var stringlist = lists.List{
	ListID:      listID,
	Description: listDescription,
	Entries:     []string{listEntries},
	Name:        listName,
	Type:        listType,
	CreatedAt:   testutil.Date,
	UpdatedAt:   testutil.Date,
	Scope: lists.Scope{
		Type: string(scope.ScopeTypeWorkspace),
	},
}

var stringlist2 = lists.List{
	ListID:      listID + "2",
	Description: listDescription + "2",
	Entries:     []string{listEntries},
	Name:        listName + "2",
	Type:        listType,
	CreatedAt:   testutil.Date,
	UpdatedAt:   testutil.Date,
	Scope: lists.Scope{
		Type: string(scope.ScopeTypeWorkspace),
	},
}

func TestIPListCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --entries flag",
			Args:      fmt.Sprintf("--name %s --workspace-id %s", listName, workspaceID),
			WantError: "error parsing arguments: required flag --entries not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--entries %s --workspace-id %s", listEntries, workspaceID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--entries %s --name %s", listEntries, listName),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--entries %s --name %s --workspace-id %s", listEntries, listName, workspaceID),
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
			Args: fmt.Sprintf("--entries %s --name %s --workspace-id %s", listEntries, listName, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(stringlist)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created Workspace IP List '%s' (list id: %s)", listName, listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--entries %s --name %s --workspace-id %s --json", listEntries, listName, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(stringlist))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(stringlist),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestIPListDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--list-id %s", listID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--list-id bar --workspace-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid List ID",
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
			Args: fmt.Sprintf("--list-id %s --workspace-id %s", listID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted Workspace IP List (list id: %s)", listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--list-id %s --workspace-id %s --json", listID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, listID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestIPListGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--list-id %s", listID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--list-id baz --workspace-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid List ID",
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
			Args: fmt.Sprintf("--list-id %s --workspace-id %s", listID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(stringlist)))),
					},
				},
			},
			WantOutput: listString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--list-id %s --workspace-id %s --json", listID, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(stringlist)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(stringlist),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestIPListList(t *testing.T) {
	listsObject := lists.Lists{
		Data: []lists.List{
			stringlist,
			stringlist2,
		},
		Meta: lists.MetaLists{},
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
			Name: "validate API success (zero workspaces)",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(lists.Lists{
							Data: []lists.List{},
							Meta: lists.MetaLists{},
						}))),
					},
				},
			},
			WantOutput: zeroListString,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--workspace-id %s", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listsObject))),
					},
				},
			},
			WantOutput: listListsString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--workspace-id %s --json", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(listsObject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(listsObject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestIPListUpdate(t *testing.T) {
	updatelist := lists.List{
		ListID:      listID,
		Description: listDescription + "2",
		Entries:     []string{listEntries + "2"},
		Name:        listName,
		Type:        listType,
		CreatedAt:   testutil.Date,
		UpdatedAt:   testutil.Date,
		Scope: lists.Scope{
			Type: string(scope.ScopeTypeWorkspace),
		},
	}
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      fmt.Sprintf("--workspace-id %s", workspaceID),
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--list-id %s", listID),
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--list-id %s --description %s --entries %s --workspace-id %s", listID, listDescription+"2", listEntries+"2", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatelist))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated Workspace IP List '%s' (list id: %s)", listName, listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--list-id %s --description %s --entries %s --workspace-id %s --json", listID, listDescription+"2", listEntries+"2", workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatelist))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updatelist),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

var listListsString = strings.TrimSpace(`
ID           Name       Description    Type  Scope      Entries  Created At
someListID   listName   NGWAFCLIList   ip    workspace  1.0.0.0  2021-06-15 23:00:00 +0000 UTC
someListID2  listName2  NGWAFCLIList2  ip    workspace  1.0.0.0  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Name  Description  Type  Scope  Entries  Created At
`) + "\n"

var listString = strings.TrimSpace(`
ID: someListID
Name: listName
Description: NGWAFCLIList
Type: ip
Entries: 1.0.0.0
Scope: workspace
Updated (UTC): 2021-06-15 23:00
`)
