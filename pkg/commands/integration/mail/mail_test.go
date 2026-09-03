package mail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	sub "github.com/fastly/cli/pkg/commands/integration/mail"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
		address         = "alerts@example.com"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--address %s", address),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args:      fmt.Sprintf("--name %s", integrationName),
			WantError: "error parsing arguments: required flag --address not provided",
		},
		{
			Args: fmt.Sprintf("--name %s --address %s", integrationName, address),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, _ *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s --address %s", integrationName, address),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, i *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					if fastly.ToValue(i.Type) != sub.CommandName {
						return nil, fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["address"] != address {
						return nil, fmt.Errorf("unexpected address: %s", i.Config["address"])
					}
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.Success("Created Mailing List integration '%s' (id: %s)", integrationName, integrationID),
		},
		{
			Args: fmt.Sprintf("--name %s --address %s --json", integrationName, address),
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
		address       = "alerts@example.com"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--address %s", address),
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args:      integrationID,
			WantError: "error parsing arguments: required flag --address not provided",
		},
		{
			Args: fmt.Sprintf("%s --address %s", integrationID, address),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, _ *fastly.UpdateIntegrationInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("%s --address %s", integrationID, address),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, i *fastly.UpdateIntegrationInput) error {
					if i.ID != integrationID {
						return fmt.Errorf("unexpected id: %s", i.ID)
					}
					if fastly.ToValue(i.Type) != sub.CommandName {
						return fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["address"] != address {
						return fmt.Errorf("unexpected address: %s", i.Config["address"])
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated Mailing List integration (id: %s)", integrationID),
		},
		{
			Args: fmt.Sprintf("%s --address %s --json", integrationID, address),
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

func TestConfirmCommand(t *testing.T) {
	const address = "alerts@example.com"

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required argument 'address' not provided",
		},
		{
			Args: address,
			API: &mock.API{
				CreateMailinglistConfirmationFn: func(_ context.Context, _ *fastly.CreateMailinglistConfirmationInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: address,
			API: &mock.API{
				CreateMailinglistConfirmationFn: func(_ context.Context, i *fastly.CreateMailinglistConfirmationInput) error {
					if fastly.ToValue(i.Email) != address {
						return fmt.Errorf("unexpected address: %s", fastly.ToValue(i.Email))
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Sent confirmation email to '%s'", address),
		},
		{
			Args: fmt.Sprintf("%s --json", address),
			API: &mock.API{
				CreateMailinglistConfirmationFn: func(_ context.Context, _ *fastly.CreateMailinglistConfirmationInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"address": %q, "sent": true}`, address),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "confirm"}, scenarios)
}
