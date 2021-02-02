package service_test

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

func TestServiceCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service", "create"},
			api:       mock.API{CreateServiceFn: createServiceOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       []string{"service", "create", "--name", "Foo"},
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       []string{"service", "create", "-n=Foo"},
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       []string{"service", "create", "--name", "Foo", "--type", "wasm"},
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       []string{"service", "create", "--name", "Foo", "--type", "wasm", "--comment", "Hello"},
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       []string{"service", "create", "-n", "Foo", "--comment", "Hello"},
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:      []string{"service", "create", "-n", "Foo"},
			api:       mock.API{CreateServiceFn: createServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestServiceList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"service", "list"},
			api:        mock.API{ListServicesFn: listServicesOK},
			wantOutput: listServicesShortOutput,
		},
		{
			args:       []string{"service", "list", "--verbose"},
			api:        mock.API{ListServicesFn: listServicesOK},
			wantOutput: listServicesVerboseOutput,
		},
		{
			args:       []string{"service", "list", "-v"},
			api:        mock.API{ListServicesFn: listServicesOK},
			wantOutput: listServicesVerboseOutput,
		},
		{
			args:       []string{"service", "--verbose", "list"},
			api:        mock.API{ListServicesFn: listServicesOK},
			wantOutput: listServicesVerboseOutput,
		},
		{
			args:       []string{"-v", "service", "list"},
			api:        mock.API{ListServicesFn: listServicesOK},
			wantOutput: listServicesVerboseOutput,
		},
		{
			args:      []string{"service", "list"},
			api:       mock.API{ListServicesFn: listServicesError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestServiceDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service", "describe"},
			api:       mock.API{GetServiceDetailsFn: describeServiceOK},
			wantError: "error reading service: no service ID found",
		},
		{
			args:       []string{"service", "describe", "--service-id", "123"},
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceShortOutput,
		},
		{
			args:       []string{"service", "describe", "--service-id", "123", "--verbose"},
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       []string{"service", "describe", "--service-id", "123", "-v"},
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       []string{"service", "--verbose", "describe", "--service-id", "123"},
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       []string{"-v", "service", "describe", "--service-id", "123"},
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:      []string{"service", "describe", "--service-id", "123"},
			api:       mock.API{GetServiceDetailsFn: describeServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
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

func TestServiceSearch(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"service", "search", "--name", "Foo"},
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceShortOutput,
		},
		{
			args:       []string{"service", "search", "--name", "Foo", "-v"},
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceVerboseOutput,
		},
		{
			args:      []string{"service", "search", "--name"},
			api:       mock.API{SearchServiceFn: searchServiceOK},
			wantError: "error parsing arguments: expected argument for flag '--name'",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
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

func TestServiceUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"service", "update"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantError: "error reading service: no service ID found",
		},
		{
			args: []string{"service", "update", "--service-id", "12345"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantError: "error parsing arguments: must provide either --name or --comment to update service",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "--name", "Foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantOutput: "Updated service 12345",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "-n=Foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantOutput: "Updated service 12345",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "--name", "Foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantOutput: "Updated service 12345",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "--name", "Foo", "--comment", "Hello"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantOutput: "Updated service 12345",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "-n", "Foo", "--comment", "Hello"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantOutput: "Updated service 12345",
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "-n", "Foo"},
			api: mock.API{
				GetServiceFn:    getServiceError,
				UpdateServiceFn: updateServiceOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"service", "update", "--service-id", "12345", "-n", "Foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestServiceDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service", "delete"},
			api:       mock.API{DeleteServiceFn: deleteServiceOK},
			wantError: "error reading service: no service ID found",
		},
		{
			args:       []string{"service", "delete", "--service-id=X"},
			api:        mock.API{DeleteServiceFn: deleteServiceOK},
			wantOutput: "Deleted service ID X",
		},
		{
			args:       []string{"service", "delete", "--service-id", "zzz"},
			api:        mock.API{DeleteServiceFn: deleteServiceOK},
			wantOutput: "Deleted service ID zzz",
		},
		{
			args:      []string{"service", "delete", "--service-id", "bonk"},
			api:       mock.API{DeleteServiceFn: deleteServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createServiceOK(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:      "12345",
		Name:    i.Name,
		Type:    i.Type,
		Comment: i.Comment,
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

func listServicesOK(i *fastly.ListServicesInput) ([]*fastly.Service, error) {
	return []*fastly.Service{
		{
			ID:            "123",
			Name:          "Foo",
			Type:          "wasm",
			CustomerID:    "mycustomerid",
			ActiveVersion: 2,
			UpdatedAt:     testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
			Versions: []*fastly.Version{
				{
					Number:    1,
					Comment:   "a",
					ServiceID: "b",
					CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
					UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
					DeletedAt: testutil.MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
				},
				{
					Number:    2,
					Comment:   "c",
					ServiceID: "d",
					Active:    true,
					Deployed:  true,
					CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
					UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
				},
			},
		},
		{
			ID:            "456",
			Name:          "Bar",
			Type:          "wasm",
			CustomerID:    "mycustomerid",
			ActiveVersion: 1,
			UpdatedAt:     testutil.MustParseTimeRFC3339("2015-03-14T12:59:59Z"),
		},
		{
			ID:            "789",
			Name:          "Baz",
			Type:          "",
			CustomerID:    "mycustomerid",
			ActiveVersion: 1,
			// nil UpdatedAt
		},
	}, nil
}

func listServicesError(i *fastly.ListServicesInput) ([]*fastly.Service, error) {
	return nil, errTest
}

var listServicesShortOutput = strings.TrimSpace(`
NAME  ID   TYPE  ACTIVE VERSION  LAST EDITED (UTC)
Foo   123  wasm  2               2010-11-15 19:01
Bar   456  wasm  1               2015-03-14 12:59
Baz   789  vcl   1               n/a
`) + "\n"

var listServicesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service 1/3
	ID: 123
	Name: Foo
	Type: wasm
	Customer ID: mycustomerid
	Last edited (UTC): 2010-11-15 19:01
	Active version: 2
	Versions: 2
		Version 1/2
			Number: 1
			Comment: a
			Service ID: b
			Active: false
			Locked: false
			Deployed: false
			Staging: false
			Testing: false
			Created (UTC): 2001-02-03 04:05
			Last edited (UTC): 2001-02-04 04:05
			Deleted (UTC): 2001-02-05 04:05
		Version 2/2
			Number: 2
			Comment: c
			Service ID: d
			Active: true
			Locked: false
			Deployed: true
			Staging: false
			Testing: false
			Created (UTC): 2001-03-03 04:05
			Last edited (UTC): 2001-03-04 04:05

Service 2/3
	ID: 456
	Name: Bar
	Type: wasm
	Customer ID: mycustomerid
	Last edited (UTC): 2015-03-14 12:59
	Active version: 1
	Versions: 0

Service 3/3
	ID: 789
	Name: Baz
	Type: vcl
	Customer ID: mycustomerid
	Active version: 1
	Versions: 0
`) + "\n\n"

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:      "12345",
		Name:    "Foo",
		Comment: "Bar",
	}, nil
}

func getServiceError(*fastly.GetServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

func describeServiceOK(i *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return &fastly.ServiceDetail{
		ID:         "123",
		Name:       "Foo",
		Type:       "wasm",
		CustomerID: "mycustomerid",
		ActiveVersion: fastly.Version{
			Number:    2,
			Comment:   "c",
			ServiceID: "d",
			Active:    true,
			Deployed:  true,
			CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
			UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
		},
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		Versions: []*fastly.Version{
			{
				Number:    1,
				Comment:   "a",
				ServiceID: "b",
				CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
				DeletedAt: testutil.MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
			},
			{
				Number:    2,
				Comment:   "c",
				ServiceID: "d",
				Active:    true,
				Deployed:  true,
				CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
			},
		},
	}, nil
}

func describeServiceError(i *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return nil, errTest
}

var describeServiceShortOutput = strings.TrimSpace(`
ID: 123
Name: Foo
Type: wasm
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version:
	Number: 2
	Comment: c
	Service ID: d
	Active: true
	Locked: false
	Deployed: true
	Staging: false
	Testing: false
	Created (UTC): 2001-03-03 04:05
	Last edited (UTC): 2001-03-04 04:05
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Locked: false
		Deployed: true
		Staging: false
		Testing: false
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

var describeServiceVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
ID: 123
Name: Foo
Type: wasm
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version:
	Number: 2
	Comment: c
	Service ID: d
	Active: true
	Locked: false
	Deployed: true
	Staging: false
	Testing: false
	Created (UTC): 2001-03-03 04:05
	Last edited (UTC): 2001-03-04 04:05
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Locked: false
		Deployed: true
		Staging: false
		Testing: false
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

func searchServiceOK(i *fastly.SearchServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:         "123",
		Name:       "Foo",
		Type:       "wasm",
		CustomerID: "mycustomerid",
		UpdatedAt:  testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		Versions: []*fastly.Version{
			{
				Number:    1,
				Comment:   "a",
				ServiceID: "b",
				CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
				DeletedAt: testutil.MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
			},
			{
				Number:    2,
				Comment:   "c",
				ServiceID: "d",
				Active:    true,
				Deployed:  true,
				CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
			},
		},
	}, nil
}

var searchServiceShortOutput = strings.TrimSpace(`
ID: 123
Name: Foo
Type: wasm
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version: 0
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Locked: false
		Deployed: true
		Staging: false
		Testing: false
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

var searchServiceVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
ID: 123
Name: Foo
Type: wasm
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version: 0
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Locked: false
		Deployed: true
		Staging: false
		Testing: false
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

func updateServiceOK(i *fastly.UpdateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:      "12345",
		Name:    *i.Name,
		Comment: *i.Comment,
	}, nil
}

func updateServiceError(*fastly.UpdateServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

func deleteServiceOK(*fastly.DeleteServiceInput) error {
	return nil
}

func deleteServiceError(*fastly.DeleteServiceInput) error {
	return errTest
}
