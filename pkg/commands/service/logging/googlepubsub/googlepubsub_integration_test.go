package googlepubsub_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/googlepubsub"
)

func TestGooglePubSubCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --topic topic --account-name=me@fastly.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreatePubsubFn: createGooglePubSubOK,
			},
			WantOutput: "Created Google Cloud Pub/Sub logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --topic topic --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreatePubsubFn: createGooglePubSubError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestGooglePubSubList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			WantOutput: listGooglePubSubsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			WantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			WantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestGooglePubSubDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetPubsubFn:    getGooglePubSubError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetPubsubFn:    getGooglePubSubOK,
			},
			WantOutput: describeGooglePubSubOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestGooglePubSubUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdatePubsubFn: updateGooglePubSubError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdatePubsubFn: updateGooglePubSubOK,
			},
			WantOutput: "Updated Google Cloud Pub/Sub logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestGooglePubSubDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeletePubsubFn: deleteGooglePubSubError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeletePubsubFn: deleteGooglePubSubOK,
			},
			WantOutput: "Deleted Google Cloud Pub/Sub logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createGooglePubSubOK(_ context.Context, i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Topic:             fastly.ToPointer("topic"),
		User:              fastly.ToPointer("user"),
		SecretKey:         fastly.ToPointer("secret"),
		ProjectID:         fastly.ToPointer("project"),
		AccountName:       fastly.ToPointer("me@fastly.com"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func createGooglePubSubError(_ context.Context, _ *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func listGooglePubSubsOK(_ context.Context, i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return []*fastly.Pubsub{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			User:              fastly.ToPointer("user@example.com"),
			AccountName:       fastly.ToPointer("none"),
			SecretKey:         fastly.ToPointer("secret"),
			ProjectID:         fastly.ToPointer("project"),
			Topic:             fastly.ToPointer("topic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			Placement:         fastly.ToPointer("none"),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			User:              fastly.ToPointer("user@example.com"),
			AccountName:       fastly.ToPointer("none"),
			SecretKey:         fastly.ToPointer("secret"),
			ProjectID:         fastly.ToPointer("project"),
			Topic:             fastly.ToPointer("analytics"),
			Placement:         fastly.ToPointer("none"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listGooglePubSubsError(_ context.Context, _ *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return nil, errTest
}

var listGooglePubSubsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGooglePubSubsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Google Cloud Pub/Sub 1/2
		Service ID: 123
		Version: 1
		Name: logs
		User: user@example.com
		Account name: none
		Secret key: secret
		Project ID: project
		Topic: topic
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Google Cloud Pub/Sub 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		User: user@example.com
		Account name: none
		Secret key: secret
		Project ID: project
		Topic: analytics
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getGooglePubSubOK(_ context.Context, i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Topic:             fastly.ToPointer("topic"),
		User:              fastly.ToPointer("user@example.com"),
		AccountName:       fastly.ToPointer("none"),
		SecretKey:         fastly.ToPointer("secret"),
		ProjectID:         fastly.ToPointer("project"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getGooglePubSubError(_ context.Context, _ *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

var describeGooglePubSubOutput = "\n" + strings.TrimSpace(`
Account name: none
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Project ID: project
Response condition: Prevent default logging
Secret key: secret
Service ID: 123
Topic: topic
User: user@example.com
Version: 1
`) + "\n"

func updateGooglePubSubOK(_ context.Context, i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Topic:             fastly.ToPointer("topic"),
		User:              fastly.ToPointer("user@example.com"),
		SecretKey:         fastly.ToPointer("secret"),
		ProjectID:         fastly.ToPointer("project"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateGooglePubSubError(_ context.Context, _ *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func deleteGooglePubSubOK(_ context.Context, _ *fastly.DeletePubsubInput) error {
	return nil
}

func deleteGooglePubSubError(_ context.Context, _ *fastly.DeletePubsubInput) error {
	return errTest
}
