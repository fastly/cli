package jiraissue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/integration"
	sub "github.com/fastly/cli/pkg/commands/integration/jiraissue"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const requiredFlags = "--base-url https://example.atlassian.net --username user@example.com --api-token abc123 --project-key PROJ --issue-type Bug"

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
					if fastly.ToValue(i.Type) != fastly.IntegrationTypeJiraIssue {
						return nil, fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					if i.Config["baseurl"] != "https://example.atlassian.net" || i.Config["username"] != "user@example.com" ||
						i.Config["token"] != "abc123" || i.Config["projectkey"] != "PROJ" || i.Config["issuetype"] != "Bug" {
						return nil, fmt.Errorf("unexpected config: %+v", i.Config)
					}
					return &fastly.CreateIntegrationResponse{ID: fastly.ToPointer(integrationID)}, nil
				},
			},
			WantOutput: fstfmt.Success("Created Jira Issue integration '%s' (id: %s)", integrationName, integrationID),
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
					if fastly.ToValue(i.Type) != fastly.IntegrationTypeJiraIssue {
						return fmt.Errorf("unexpected type: %s", fastly.ToValue(i.Type))
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Updated Jira Issue integration (id: %s)", integrationID),
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
