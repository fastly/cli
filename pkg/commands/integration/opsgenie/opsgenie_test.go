package opsgenie_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	sub "github.com/fastly/cli/pkg/commands/integration/opsgenie"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
		apiKey          = "a1b2c3d4"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--api-key %s", apiKey),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args:      fmt.Sprintf("--name %s", integrationName),
			WantError: "error parsing arguments: required flag --api-key not provided",
		},
		{
			Args: fmt.Sprintf("--name %s --api-key %s", integrationName, apiKey),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, _ *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s --api-key %s", integrationName, apiKey),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, i *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					if fastly.ToValue(i.Type) != fastly.IntegrationTypeOpsGenie {
						return nil, fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["apikey"] != apiKey {
						return nil, fmt.Errorf("unexpected apikey: %s", i.Config["apikey"])
					}
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.Success("Created OpsGenie integration '%s' (id: %s)", integrationName, integrationID),
		},
		{
			Args: fmt.Sprintf("--name %s --api-key %s --json", integrationName, apiKey),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, _ *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestUpdateCommand(t *testing.T) {
	const (
		integrationID = "integration-id-123"
		apiKey        = "a1b2c3d4"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--api-key %s", apiKey),
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args: fmt.Sprintf("%s --name new-name", integrationID),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, i *fastly.UpdateIntegrationInput) error {
					if i.Config != nil {
						return fmt.Errorf("unexpected config: %+v", i.Config)
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated OpsGenie integration (id: %s)", integrationID),
		},
		{
			Args: fmt.Sprintf("%s --api-key %s", integrationID, apiKey),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, _ *fastly.UpdateIntegrationInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("%s --api-key %s", integrationID, apiKey),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, i *fastly.UpdateIntegrationInput) error {
					if i.ID != integrationID {
						return fmt.Errorf("unexpected id: %s", i.ID)
					}
					if fastly.ToValue(i.Type) != fastly.IntegrationTypeOpsGenie {
						return fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["apikey"] != apiKey {
						return fmt.Errorf("unexpected apikey: %s", i.Config["apikey"])
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated OpsGenie integration (id: %s)", integrationID),
		},
		{
			Args: fmt.Sprintf("%s --api-key %s --json", integrationID, apiKey),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, _ *fastly.UpdateIntegrationInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "updated": true}`, integrationID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
