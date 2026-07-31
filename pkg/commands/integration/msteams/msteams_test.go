package msteams_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	sub "github.com/fastly/cli/pkg/commands/integration/msteams"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
		webhookURL      = "https://outlook.office.com/webhook/xyz"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--webhook %s", webhookURL),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args:      fmt.Sprintf("--name %s", integrationName),
			WantError: "error parsing arguments: required flag --webhook not provided",
		},
		{
			Args: fmt.Sprintf("--name %s --webhook %s", integrationName, webhookURL),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, _ *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s --webhook %s", integrationName, webhookURL),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, i *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					if fastly.ToValue(i.Type) != sub.CommandName {
						return nil, fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["webhook"] != webhookURL {
						return nil, fmt.Errorf("unexpected webhook: %s", i.Config["webhook"])
					}
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.Success("Created Microsoft Teams integration '%s' (id: %s)", integrationName, integrationID),
		},
		{
			Args: fmt.Sprintf("--name %s --webhook %s --json", integrationName, webhookURL),
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
		webhookURL    = "https://outlook.office.com/webhook/xyz"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--webhook %s", webhookURL),
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args:      integrationID,
			WantError: "error parsing arguments: required flag --webhook not provided",
		},
		{
			Args: fmt.Sprintf("%s --webhook %s", integrationID, webhookURL),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, _ *fastly.UpdateIntegrationInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("%s --webhook %s", integrationID, webhookURL),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, i *fastly.UpdateIntegrationInput) error {
					if i.ID != integrationID {
						return fmt.Errorf("unexpected id: %s", i.ID)
					}
					if fastly.ToValue(i.Type) != sub.CommandName {
						return fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["webhook"] != webhookURL {
						return fmt.Errorf("unexpected webhook: %s", i.Config["webhook"])
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated Microsoft Teams integration (id: %s)", integrationID),
		},
		{
			Args: fmt.Sprintf("%s --webhook %s --json", integrationID, webhookURL),
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
