package stringlist_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/stringlist"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
)

const (
	listID          = "someListID"
	listDescription = "NGWAFCLIList"
	listEntries     = "1.0.0.0"
	listType        = "string"
	listName        = "listName"
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
		Type: string(scope.ScopeTypeAccount),
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
		Type: string(scope.ScopeTypeAccount),
	},
}

func TestStringListCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --entries flag",
			Args:      fmt.Sprintf("--name %s", listName),
			WantError: "error parsing arguments: required flag --entries not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      fmt.Sprintf("--entries %s", listEntries),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--entries %s --name %s", listEntries, listName),
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
			Args: fmt.Sprintf("--entries %s --name %s", listEntries, listName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(stringlist)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created Account String List '%s' (list id: %s)", listName, listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--entries %s --name %s --json", listEntries, listName),
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

func TestStringListDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--list-id bar",
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
			Args: fmt.Sprintf("--list-id %s", listID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted Account String List (list id: %s)", listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--list-id %s --json", listID),
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

func TestStringListGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--list-id baz",
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
			Args: fmt.Sprintf("--list-id %s", listID),
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
			Args: fmt.Sprintf("--list-id %s --json", listID),
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

func TestStringListList(t *testing.T) {
	listsObject := lists.Lists{
		Data: []lists.List{
			stringlist,
			stringlist2,
		},
		Meta: lists.MetaLists{},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate internal server error",
			Args: "",
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
			Args: "",
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
			Args: "",
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
			Args: "--json",
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

func TestStringListUpdate(t *testing.T) {
	updatelist := lists.List{
		ListID:      listID,
		Description: listDescription + "2",
		Entries:     []string{listEntries + "2"},
		Name:        listName,
		Type:        listType,
		CreatedAt:   testutil.Date,
		UpdatedAt:   testutil.Date,
		Scope: lists.Scope{
			Type: string(scope.ScopeTypeAccount),
		},
	}
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --list-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --list-id not provided",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--list-id %s --description %s --entries %s", listID, listDescription+"2", listEntries+"2"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updatelist))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated Account String List '%s' (list id: %s)", listName, listID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--list-id %s --description %s --entries %s --json", listID, listDescription+"2", listEntries+"2"),
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
ID           Name       Description    Type    Scope    Entries  Updated At                     Created At
someListID   listName   NGWAFCLIList   string  account  1.0.0.0  2021-06-15 23:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
someListID2  listName2  NGWAFCLIList2  string  account  1.0.0.0  2021-06-15 23:00:00 +0000 UTC  2021-06-15 23:00:00 +0000 UTC
`) + "\n"

var zeroListString = strings.TrimSpace(`
ID  Name  Description  Type  Scope  Entries  Updated At  Created At
`) + "\n"

var listString = strings.TrimSpace(`
ID: someListID
Name: listName
Description: NGWAFCLIList
Type: string
Entries: 1.0.0.0
Scope: account
Updated (UTC): 2021-06-15 23:00
`)
