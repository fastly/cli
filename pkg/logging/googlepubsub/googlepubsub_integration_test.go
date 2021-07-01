package googlepubsub_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestGooglePubSubCreate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --secret-key secret --project-id project --topic topic --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --project-id project --topic topic --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --secret-key secret --topic topic --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --project-id not provided",
		},
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --topic not provided",
		},
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --topic topic --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreatePubsubFn: createGooglePubSubOK,
			},
			wantOutput: "Created Google Cloud Pub/Sub logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --topic topic --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreatePubsubFn: createGooglePubSubError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubList(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging googlepubsub list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			wantOutput: listGooglePubSubsShortOutput,
		},
		{
			args: args("logging googlepubsub list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args: args("logging googlepubsub list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args: args("logging googlepubsub --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args: args("logging -v googlepubsub list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsOK,
			},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args: args("logging googlepubsub list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListPubsubsFn:  listGooglePubSubsError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestGooglePubSubDescribe(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging googlepubsub describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging googlepubsub describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetPubsubFn:    getGooglePubSubError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging googlepubsub describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetPubsubFn:    getGooglePubSubOK,
			},
			wantOutput: describeGooglePubSubOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestGooglePubSubUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging googlepubsub update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging googlepubsub update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdatePubsubFn: updateGooglePubSubError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging googlepubsub update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdatePubsubFn: updateGooglePubSubOK,
			},
			wantOutput: "Updated Google Cloud Pub/Sub logging endpoint log (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubDelete(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging googlepubsub delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging googlepubsub delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeletePubsubFn: deleteGooglePubSubError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging googlepubsub delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeletePubsubFn: deleteGooglePubSubOK,
			},
			wantOutput: "Deleted Google Cloud Pub/Sub logging endpoint logs (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createGooglePubSubOK(i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Topic:             "topic",
		User:              "user",
		SecretKey:         "secret",
		ProjectID:         "project",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func createGooglePubSubError(i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func listGooglePubSubsOK(i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return []*fastly.Pubsub{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			User:              "user@example.com",
			SecretKey:         "secret",
			ProjectID:         "project",
			Topic:             "topic",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			Placement:         "none",
			FormatVersion:     2,
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			User:              "user@example.com",
			SecretKey:         "secret",
			ProjectID:         "project",
			Topic:             "analytics",
			Placement:         "none",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
		},
	}, nil
}

func listGooglePubSubsError(i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return nil, errTest
}

var listGooglePubSubsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGooglePubSubsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Google Cloud Pub/Sub 1/2
		Service ID: 123
		Version: 1
		Name: logs
		User: user@example.com
		Secret key: secret
		Project ID: project
		Topic: topic
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Google Cloud Pub/Sub 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		User: user@example.com
		Secret key: secret
		Project ID: project
		Topic: analytics
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getGooglePubSubOK(i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Topic:             "topic",
		User:              "user@example.com",
		SecretKey:         "secret",
		ProjectID:         "project",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getGooglePubSubError(i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

var describeGooglePubSubOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
User: user@example.com
Secret key: secret
Project ID: project
Topic: topic
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateGooglePubSubOK(i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Topic:             "topic",
		User:              "user@example.com",
		SecretKey:         "secret",
		ProjectID:         "project",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateGooglePubSubError(i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func deleteGooglePubSubOK(i *fastly.DeletePubsubInput) error {
	return nil
}

func deleteGooglePubSubError(i *fastly.DeletePubsubInput) error {
	return errTest
}
