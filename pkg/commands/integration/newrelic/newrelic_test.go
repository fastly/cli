package newrelic_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	sub "github.com/fastly/cli/pkg/commands/integration/newrelic"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const requiredFlags = "--account-id acct123 --api-key key123"

func TestCreateCommand(t *testing.T) {
	const (
		integrationName = "test123"
		integrationID   = "integration-id-123"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      requiredFlags,
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: fmt.Sprintf("--name %s %s", integrationName, requiredFlags),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, _ *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s %s", integrationName, requiredFlags),
			API: &mock.API{
				CreateIntegrationFn: func(_ context.Context, i *fastly.CreateIntegrationInput) (*fastly.CreateIntegrationResponse, error) {
					if fastly.ToValue(i.Type) != sub.CommandName {
						return nil, fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["account"] != "acct123" || i.Config["key"] != "key123" {
						return nil, fmt.Errorf("unexpected config: %+v", i.Config)
					}
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.Success("Created New Relic integration '%s' (id: %s)", integrationName, integrationID),
		},
		{
			Args: fmt.Sprintf("--name %s %s --json", integrationName, requiredFlags),
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
	const integrationID = "integration-id-123"

	scenarios := []testutil.CLIScenario{
		{
			Args:      requiredFlags,
			WantError: "error parsing arguments: required argument 'id' not provided",
		},
		{
			Args: fmt.Sprintf("%s %s", integrationID, requiredFlags),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, _ *fastly.UpdateIntegrationInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("%s %s", integrationID, requiredFlags),
			API: &mock.API{
				UpdateIntegrationFn: func(_ context.Context, i *fastly.UpdateIntegrationInput) error {
					if i.ID != integrationID {
						return fmt.Errorf("unexpected id: %s", i.ID)
					}
					if fastly.ToValue(i.Type) != sub.CommandName {
						return fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated New Relic integration (id: %s)", integrationID),
		},
		{
			Args: fmt.Sprintf("%s %s --json", integrationID, requiredFlags),
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
