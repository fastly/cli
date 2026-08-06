package eventmapping_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/auditlog"
	sub "github.com/fastly/cli/pkg/commands/auditlog/eventmapping"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings/eventtypes"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings/scopetypes"
)

const (
	mappingID   = "mappingID"
	mappingName = "mappingName"
)

var em = eventmappings.EventMapping{
	ID:             mappingID,
	CustomerID:     "customerID",
	Name:           mappingName,
	Description:    "mappingDescription",
	ScopeType:      eventmappings.ScopeTypeAccount,
	ScopeIDs:       []string{},
	EventTypes:     []string{"user.login"},
	IntegrationIDs: []string{"integrationID"},
	MappingStatus:  eventmappings.MappingStatusActive,
	CreatedAt:      testutil.Date,
	UpdatedAt:      testutil.Date,
}

func TestEventMappingCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--scope-type account --event-type user.login --integration-id integrationID",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --scope-type flag",
			Args:      "--name foo --event-type user.login --integration-id integrationID",
			WantError: "error parsing arguments: required flag --scope-type not provided",
		},
		{
			Name:      "validate missing --event-type flag",
			Args:      "--name foo --scope-type account --integration-id integrationID",
			WantError: "error parsing arguments: required flag --event-type not provided",
		},
		{
			Name:      "validate missing --integration-id flag",
			Args:      "--name foo --scope-type account --event-type user.login",
			WantError: "error parsing arguments: required flag --integration-id not provided",
		},
		{
			Name: "validate internal server error",
			Args: "--name foo --scope-type account --event-type user.login --integration-id integrationID",
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
			Args: fmt.Sprintf("--name %s --scope-type %s --event-type %s --integration-id %s", mappingName, eventmappings.ScopeTypeAccount, "user.login", "integrationID"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created event mapping '%s' (id: %s)", em.Name, em.ID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --scope-type %s --event-type %s --integration-id %s --json", mappingName, eventmappings.ScopeTypeAccount, "user.login", "integrationID"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(em),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestEventMappingDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--id %s", mappingID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted event mapping (id: %s)", mappingID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--id %s --json", mappingID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, mappingID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestEventMappingDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--id %s", mappingID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: emString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--id %s --json", mappingID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(em),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestEventMappingList(t *testing.T) {
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
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(eventmappings.Collection{
							Data: []eventmappings.EventMapping{em},
							Meta: eventmappings.Meta{Total: 1},
						}))),
					},
				},
			},
			WantOutput: mappingID,
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(eventmappings.Collection{
							Data: []eventmappings.EventMapping{em},
							Meta: eventmappings.Meta{Total: 1},
						}))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON([]eventmappings.EventMapping{em}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestEventMappingUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "--name foo --scope-type account --event-type user.login --integration-id integrationID",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--id %s --name foo --scope-type account --event-type user.login --integration-id integrationID", mappingID),
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
			Args: fmt.Sprintf("--id %s --name %s --scope-type %s --event-type %s --integration-id %s", mappingID, mappingName, eventmappings.ScopeTypeAccount, "user.login", "integrationID"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated event mapping '%s' (id: %s)", em.Name, em.ID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--id %s --name %s --scope-type %s --event-type %s --integration-id %s --json", mappingID, mappingName, eventmappings.ScopeTypeAccount, "user.login", "integrationID"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(em))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(em),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestEventMappingListEventTypes(t *testing.T) {
	types := eventtypes.Collection{
		Data: []eventtypes.EventType{
			{EventType: "user.login", DisplayName: "User Login", ScopeTypes: []string{"account"}},
		},
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
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(types))),
					},
				},
			},
			WantOutput: "user.login",
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(types))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(types),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-event-types"}, scenarios)
}

func TestEventMappingListScopeTypes(t *testing.T) {
	types := scopetypes.Collection{
		Data: []scopetypes.ScopeType{
			{ScopeType: "account"},
		},
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
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(types))),
					},
				},
			},
			WantOutput: "account",
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(types))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(types),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-scope-types"}, scenarios)
}

var emString = fmt.Sprintf(`ID: %s
Name: %s
Description: %s
Scope Type: %s
Scope IDs: 
Event Types: %s
Integration IDs: %s
Mapping Status: %s
Created (UTC): %s
Updated (UTC): %s
`, em.ID, em.Name, em.Description, em.ScopeType, em.EventTypes[0], em.IntegrationIDs[0], em.MappingStatus,
	testutil.Date.UTC().Format("2006-01-02 15:04"), testutil.Date.UTC().Format("2006-01-02 15:04"))
