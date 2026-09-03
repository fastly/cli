package integration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestDescribeIntegrationCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
	)

	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args: integrationID,
			API: &mock.API{
				GetIntegrationFn: func(_ context.Context, _ *fastly.GetIntegrationInput) (*fastly.Integration, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: integrationID,
			API: &mock.API{
				GetIntegrationFn: func(_ context.Context, i *fastly.GetIntegrationInput) (*fastly.Integration, error) {
					return &fastly.Integration{
						ID:        &i.ID,
						Name:      fastly.ToPointer(integrationName),
						Type:      fastly.ToPointer(fastly.IntegrationTypeDatadog),
						Status:    fastly.ToPointer("enabled"),
						Config:    map[string]string{"apikey": "abc123"},
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fmtIntegration(&fastly.Integration{
				ID:        fastly.ToPointer(integrationID),
				Name:      fastly.ToPointer(integrationName),
				Type:      fastly.ToPointer(fastly.IntegrationTypeDatadog),
				Status:    fastly.ToPointer("enabled"),
				Config:    map[string]string{"apikey": "abc123"},
				CreatedAt: &now,
			}),
		},
		{
			Args: fmt.Sprintf("%s --json", integrationID),
			API: &mock.API{
				GetIntegrationFn: func(_ context.Context, i *fastly.GetIntegrationInput) (*fastly.Integration, error) {
					return &fastly.Integration{
						ID:        &i.ID,
						Name:      fastly.ToPointer(integrationName),
						Type:      fastly.ToPointer(fastly.IntegrationTypeDatadog),
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.Integration{
				ID:        fastly.ToPointer(integrationID),
				Name:      fastly.ToPointer(integrationName),
				Type:      fastly.ToPointer(fastly.IntegrationTypeDatadog),
				CreatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestDeleteIntegrationCommand(t *testing.T) {
	const integrationID = "integration-id-123"
	errNotFound := errors.New("integration not found")

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args: "DOES-NOT-EXIST",
			API: &mock.API{
				DeleteIntegrationFn: func(_ context.Context, i *fastly.DeleteIntegrationInput) error {
					if i.ID != integrationID {
						return errNotFound
					}
					return nil
				},
			},
			WantError: errNotFound.Error(),
		},
		{
			Args: integrationID,
			API: &mock.API{
				DeleteIntegrationFn: func(_ context.Context, i *fastly.DeleteIntegrationInput) error {
					if i.ID != integrationID {
						return errNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted integration '%s'\n", integrationID),
		},
		{
			Args: fmt.Sprintf("%s --json", integrationID),
			API: &mock.API{
				DeleteIntegrationFn: func(_ context.Context, i *fastly.DeleteIntegrationInput) error {
					if i.ID != integrationID {
						return errNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, integrationID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestListIntegrationsCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
	)

	now := time.Now()

	integrations := &fastly.SearchIntegrationsResponse{
		Data: []fastly.Integration{
			{ID: fastly.ToPointer(integrationID), Name: fastly.ToPointer(integrationName), Type: fastly.ToPointer(fastly.IntegrationTypeDatadog), CreatedAt: &now},
			{ID: fastly.ToPointer(integrationID + "+1"), Name: fastly.ToPointer(integrationName + "+1"), Type: fastly.ToPointer(fastly.IntegrationTypeOpsGenie), CreatedAt: &now},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			API: &mock.API{
				SearchIntegrationsFn: func(_ context.Context, _ *fastly.SearchIntegrationsInput) (*fastly.SearchIntegrationsResponse, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			API: &mock.API{
				SearchIntegrationsFn: func(_ context.Context, _ *fastly.SearchIntegrationsInput) (*fastly.SearchIntegrationsResponse, error) {
					return integrations, nil
				},
			},
			WantOutput: fmtIntegrations(integrations.Data),
		},
		{
			Args: fmt.Sprintf("--type %s", fastly.IntegrationTypeDatadog),
			API: &mock.API{
				SearchIntegrationsFn: func(_ context.Context, i *fastly.SearchIntegrationsInput) (*fastly.SearchIntegrationsResponse, error) {
					if fastly.ToValue(i.Type) != fastly.IntegrationTypeDatadog {
						return nil, errors.New("unexpected type filter")
					}
					return integrations, nil
				},
			},
			WantOutput: fmtIntegrations(integrations.Data),
		},
		{
			Args: "--json",
			API: &mock.API{
				SearchIntegrationsFn: func(_ context.Context, _ *fastly.SearchIntegrationsInput) (*fastly.SearchIntegrationsResponse, error) {
					return integrations, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(integrations),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestListTypesCommand(t *testing.T) {
	types := []fastly.IntegrationType{
		{
			Type:        fastly.ToPointer(fastly.IntegrationTypeDatadog),
			DisplayName: fastly.ToPointer("Datadog"),
			CustomFields: []fastly.CustomField{
				{Name: fastly.ToPointer("apikey"), DisplayName: fastly.ToPointer("API Key"), Format: fastly.ToPointer("string")},
			},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			API: &mock.API{
				GetIntegrationTypesFn: func(_ context.Context) (*[]fastly.IntegrationType, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			API: &mock.API{
				GetIntegrationTypesFn: func(_ context.Context) (*[]fastly.IntegrationType, error) {
					return &types, nil
				},
			},
			WantOutput: fmtIntegrationTypes(types),
		},
		{
			Args: "--json",
			API: &mock.API{
				GetIntegrationTypesFn: func(_ context.Context) (*[]fastly.IntegrationType, error) {
					return &types, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&types),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list-types"}, scenarios)
}

func fmtIntegration(i *fastly.Integration) string {
	var b bytes.Buffer
	text.PrintIntegration(&b, i)
	return b.String()
}

func fmtIntegrations(integrations []fastly.Integration) string {
	var b bytes.Buffer
	text.PrintIntegrationsTbl(&b, integrations)
	return b.String()
}

func fmtIntegrationTypes(types []fastly.IntegrationType) string {
	var b bytes.Buffer
	text.PrintIntegrationTypesTbl(&b, types)
	return b.String()
}
