package googlepubsub_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestGooglePubSubCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging googlepubsub create --service-id 123 --version 1 --name log --user user@example.com --secret-key secret --project-id project --topic topic --account-name=me@fastly.com --autoclone"),
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestGooglePubSubDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestGooglePubSubUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createGooglePubSubOK(i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
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

func createGooglePubSubError(_ *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func listGooglePubSubsOK(i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
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
		},
	}, nil
}

func listGooglePubSubsError(_ *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
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
`) + "\n\n"

func getGooglePubSubOK(i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
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
	}, nil
}

func getGooglePubSubError(_ *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

var describeGooglePubSubOutput = "\n" + strings.TrimSpace(`
Account name: none
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Project ID: project
Response condition: Prevent default logging
Secret key: secret
Service ID: 123
Topic: topic
User: user@example.com
Version: 1
`) + "\n"

func updateGooglePubSubOK(i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
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

func updateGooglePubSubError(_ *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return nil, errTest
}

func deleteGooglePubSubOK(_ *fastly.DeletePubsubInput) error {
	return nil
}

func deleteGooglePubSubError(_ *fastly.DeletePubsubInput) error {
	return errTest
}
