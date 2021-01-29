package googlepubsub_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestGooglePubSubCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--secret-key", "secret", "--project-id", "project", "--topic", "topic"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user@example.com", "--project-id", "project", "--topic", "topic"},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args:      []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user@example.com", "--secret-key", "secret", "--topic", "topic"},
			wantError: "error parsing arguments: required flag --project-id not provided",
		},
		{
			args:      []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user@example.com", "--secret-key", "secret", "--project-id", "project"},
			wantError: "error parsing arguments: required flag --topic not provided",
		},
		{
			args:       []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user@example.com", "--secret-key", "secret", "--project-id", "project", "--topic", "topic"},
			api:        mock.API{CreatePubsubFn: createGooglePubSubOK},
			wantOutput: "Created Google Cloud Pub/Sub logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "googlepubsub", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user@example.com", "--secret-key", "secret", "--project-id", "project", "--topic", "topic"},
			api:       mock.API{CreatePubsubFn: createGooglePubSubError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "googlepubsub", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPubsubsFn: listGooglePubSubsOK},
			wantOutput: listGooglePubSubsShortOutput,
		},
		{
			args:       []string{"logging", "googlepubsub", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListPubsubsFn: listGooglePubSubsOK},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args:       []string{"logging", "googlepubsub", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListPubsubsFn: listGooglePubSubsOK},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args:       []string{"logging", "googlepubsub", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPubsubsFn: listGooglePubSubsOK},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "googlepubsub", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPubsubsFn: listGooglePubSubsOK},
			wantOutput: listGooglePubSubsVerboseOutput,
		},
		{
			args:      []string{"logging", "googlepubsub", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListPubsubsFn: listGooglePubSubsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestGooglePubSubDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "googlepubsub", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "googlepubsub", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetPubsubFn: getGooglePubSubError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "googlepubsub", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetPubsubFn: getGooglePubSubOK},
			wantOutput: describeGooglePubSubOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestGooglePubSubUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "googlepubsub", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "googlepubsub", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPubsubFn:    getGooglePubSubError,
				UpdatePubsubFn: updateGooglePubSubOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "googlepubsub", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPubsubFn:    getGooglePubSubOK,
				UpdatePubsubFn: updateGooglePubSubError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "googlepubsub", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPubsubFn:    getGooglePubSubOK,
				UpdatePubsubFn: updateGooglePubSubOK,
			},
			wantOutput: "Updated Google Cloud Pub/Sub logging endpoint log (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestGooglePubSubDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "googlepubsub", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "googlepubsub", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeletePubsubFn: deleteGooglePubSubError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "googlepubsub", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeletePubsubFn: deleteGooglePubSubOK},
			wantOutput: "Deleted Google Cloud Pub/Sub logging endpoint logs (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
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
