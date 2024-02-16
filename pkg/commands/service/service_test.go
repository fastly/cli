package service_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/manifest"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
)

func TestServiceCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       args("service create --name Foo"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create -n=Foo"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create --name Foo --type wasm"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create --name Foo --type wasm --comment Hello"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create -n Foo --comment Hello"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:      args("service create -n Foo"),
			api:       mock.API{CreateServiceFn: createServiceError},
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

func TestServiceList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			api: mock.API{
				GetServicesFn: func(i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](mock.HTTPClient{
						Errors: []error{
							testutil.Err,
						},
						Responses: []*http.Response{nil},
					}, fastly.ListOpts{}, "/example")
				},
			},
			args:      args("service list"),
			wantError: testutil.Err.Error(),
		},
		{
			api: mock.API{
				GetServicesFn: func(i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[
                  {
                    "name": "Foo",
                    "id": "123",
                    "type": "wasm",
                    "version": 2,
                    "updated_at": "2021-06-15T23:00:00Z"
                  },
                  {
                    "name": "Bar",
                    "id": "456",
                    "type": "wasm",
                    "version": 1,
                    "updated_at": "2021-06-15T23:00:00Z"
                  },
                  {
                    "name": "Baz",
                    "id": "789",
                    "type": "vcl",
                    "version": 1
                  }
                ]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			args:       args("service list --per-page 1"),
			wantOutput: listServicesShortOutput,
		},
		{
			api: mock.API{
				GetServicesFn: func(i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[
                  {
                    "name": "Foo",
                    "id": "123",
                    "type": "wasm",
                    "version": 2,
                    "updated_at": "2021-06-15T23:00:00Z",
                    "customer_id": "mycustomerid",
                    "versions": [
                      {
                        "number": 1,
                        "comment": "a",
                        "service_id": "b",
                        "active": false,
                        "locked": false,
                        "deployed": false,
                        "staging": false,
                        "testing": false,
                        "created_at": "2021-06-15T23:00:00Z",
                        "deleted_at": "2021-06-15T23:00:00Z",
                        "updated_at": "2021-06-15T23:00:00Z"
                      },
                      {
                        "number": 2,
                        "comment": "c",
                        "service_id": "d",
                        "active": true,
                        "locked": false,
                        "deployed": true,
                        "staging": false,
                        "testing": false,
                        "created_at": "2021-06-15T23:00:00Z",
                        "updated_at": "2021-06-15T23:00:00Z"
                      }
                    ]
                  },
                  {
                    "name": "Bar",
                    "id": "456",
                    "type": "wasm",
                    "version": 1,
                    "updated_at": "2021-06-15T23:00:00Z",
                    "customer_id": "mycustomerid"
                  },
                  {
                    "name": "Baz",
                    "id": "789",
                    "type": "vcl",
                    "version": 1,
                    "customer_id": "mycustomerid"
                  }
                ]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			args:       args("service list --verbose"),
			wantOutput: listServicesVerboseOutput,
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

func TestServiceDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service describe"),
			api:       mock.API{GetServiceDetailsFn: describeServiceOK},
			wantError: "error reading service: no service ID found",
		},
		{
			args:       args("service describe --service-id 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceShortOutput,
		},
		{
			args:       args("service describe --service-id 123 --verbose"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("service describe --service-id 123 -v"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("service --verbose describe --service-id 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("-v service describe --service-id 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:      args("service describe --service-id 123"),
			api:       mock.API{GetServiceDetailsFn: describeServiceError},
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

func TestServiceSearch(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service search"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       args("service search --name Foo"),
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceShortOutput,
		},
		{
			args:       args("service search --name Foo -v"),
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceVerboseOutput,
		},
		{
			args:      args("service search --name"),
			api:       mock.API{SearchServiceFn: searchServiceOK},
			wantError: "error parsing arguments: expected argument for flag '--name'",
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

func TestServiceUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("service update"),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantError: "error reading service: no service ID found",
		},
		{
			args:      args("service update --service-id 12345"),
			api:       mock.API{UpdateServiceFn: updateServiceOK},
			wantError: "error parsing arguments: must provide either --name or --comment to update service",
		},
		{
			args:       args("service update --service-id 12345 --name Foo"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 -n=Foo"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 --name Foo"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 --name Foo --comment Hello"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 -n Foo --comment Hello"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:      args("service update --service-id 12345 -n Foo"),
			api:       mock.API{UpdateServiceFn: updateServiceError},
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

func TestServiceDelete(t *testing.T) {
	args := testutil.Args
	nonEmptyServiceID := regexp.MustCompile(`service_id = "[^"]+"`)

	scenarios := []struct {
		args                 []string
		api                  mock.API
		manifest             string
		wantError            string
		wantOutput           string
		expectEmptyServiceID bool
	}{
		{
			args:      args("service delete"),
			api:       mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:  "fastly-no-serviceid.toml",
			wantError: "error reading service: no service ID found",
		},
		{
			args:                 args("service delete"),
			api:                  mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:             "fastly-valid.toml",
			wantOutput:           "Deleted service ID 123",
			expectEmptyServiceID: true,
		},
		{
			args:       args("service delete --service-id 001"),
			api:        mock.API{DeleteServiceFn: deleteServiceOK},
			wantOutput: "Deleted service ID 001",
		},
		{
			args:                 args("service delete --service-id 001"),
			api:                  mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:             "fastly-valid.toml",
			wantOutput:           "Deleted service ID 001",
			expectEmptyServiceID: false,
		},
		{
			args:      args("service delete --service-id 001"),
			api:       mock.API{DeleteServiceFn: deleteServiceError},
			manifest:  "fastly-valid.toml",
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			// We're going to chdir to an temp environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			opts := testutil.EnvOpts{T: t}
			if testcase.manifest != "" {
				b, err := os.ReadFile(filepath.Join("testdata", testcase.manifest))
				if err != nil {
					t.Fatal(err)
				}
				opts.Write = []testutil.FileIO{
					{Src: string(b), Dst: manifest.Filename},
				}
			}
			rootdir := testutil.NewEnv(opts)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the temp environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(pwd)
			}()

			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				runOpts := testutil.MockGlobalData(testcase.args, &stdout)
				runOpts.APIClientFactory = mock.APIClient(testcase.api)
				return runOpts, nil
			}
			runErr := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, runErr, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)

			if testcase.manifest != "" {
				m := filepath.Join(rootdir, manifest.Filename)
				b, err := os.ReadFile(m)
				if err != nil {
					t.Fatal(err)
				}

				if testcase.expectEmptyServiceID {
					testutil.AssertStringContains(t, string(b), `service_id = ""`)
				} else if !nonEmptyServiceID.Match(b) && runErr == nil {
					// The runErr check is to prevent the first test case from causing an
					// accidental failure. As the fastly.toml doesn't have a service_id
					// set, while marshalling back and forth it'll get converted to an
					// empty string in the manifest file which will accidentally trigger
					// the following test error otherwise if we don't check for the nil
					// error value. Because that first test case expects an error to be
					// raised we know that we can safely check for `runErr == nil` here.
					t.Fatal("expected service_id to contain a value")
				}
			}
		})
	}
}

var errTest = errors.New("fixture error")

func createServiceOK(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ServiceID: fastly.ToPointer("12345"),
		Name:      i.Name,
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

var listServicesShortOutput = strings.TrimSpace(`
NAME  ID   TYPE  ACTIVE VERSION  LAST EDITED (UTC)
Foo   123  wasm  2               2021-06-15 23:00
Bar   456  wasm  1               2021-06-15 23:00
Baz   789  vcl   1               n/a
`) + "\n"

var listServicesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service 1/3
	ID: 123
	Name: Foo
	Type: wasm
	Customer ID: mycustomerid
	Last edited (UTC): 2021-06-15 23:00
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
			Created (UTC): 2021-06-15 23:00
			Last edited (UTC): 2021-06-15 23:00
			Deleted (UTC): 2021-06-15 23:00
		Version 2/2
			Number: 2
			Comment: c
			Service ID: d
			Active: true
			Locked: false
			Deployed: true
			Staging: false
			Testing: false
			Created (UTC): 2021-06-15 23:00
			Last edited (UTC): 2021-06-15 23:00

Service 2/3
	ID: 456
	Name: Bar
	Type: wasm
	Customer ID: mycustomerid
	Last edited (UTC): 2021-06-15 23:00
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

func getServiceOK(_ *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ServiceID: fastly.ToPointer("12345"),
		Name:      fastly.ToPointer("Foo"),
		Comment:   fastly.ToPointer("Bar"),
	}, nil
}

func describeServiceOK(_ *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return &fastly.ServiceDetail{
		ServiceID:  fastly.ToPointer("123"),
		Name:       fastly.ToPointer("Foo"),
		Type:       fastly.ToPointer("wasm"),
		Comment:    fastly.ToPointer("example"),
		CustomerID: fastly.ToPointer("mycustomerid"),
		ActiveVersion: &fastly.Version{
			Number:    fastly.ToPointer(2),
			Comment:   fastly.ToPointer("c"),
			ServiceID: fastly.ToPointer("d"),
			Active:    fastly.ToPointer(true),
			Deployed:  fastly.ToPointer(true),
			CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
			UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
		},
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		Versions: []*fastly.Version{
			{
				Number:    fastly.ToPointer(1),
				Comment:   fastly.ToPointer("a"),
				ServiceID: fastly.ToPointer("b"),
				CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
				DeletedAt: testutil.MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
			},
			{
				Number:    fastly.ToPointer(2),
				Comment:   fastly.ToPointer("c"),
				ServiceID: fastly.ToPointer("d"),
				Active:    fastly.ToPointer(true),
				Deployed:  fastly.ToPointer(true),
				CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-03-04T04:05:06Z"),
			},
		},
	}, nil
}

func describeServiceError(_ *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return nil, errTest
}

var describeServiceShortOutput = strings.TrimSpace(`
ID: 123
Name: Foo
Type: wasm
Comment: example
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version:
	Number: 2
	Comment: c
	Service ID: d
	Active: true
	Deployed: true
	Created (UTC): 2001-03-03 04:05
	Last edited (UTC): 2001-03-04 04:05
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Deployed: true
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

var describeServiceVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

ID: 123
Name: Foo
Type: wasm
Comment: example
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Active version:
	Number: 2
	Comment: c
	Service ID: d
	Active: true
	Deployed: true
	Created (UTC): 2001-03-03 04:05
	Last edited (UTC): 2001-03-04 04:05
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Deployed: true
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

func searchServiceOK(_ *fastly.SearchServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ServiceID:  fastly.ToPointer("123"),
		Name:       fastly.ToPointer("Foo"),
		Type:       fastly.ToPointer("wasm"),
		CustomerID: fastly.ToPointer("mycustomerid"),
		UpdatedAt:  testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		Versions: []*fastly.Version{
			{
				Number:    fastly.ToPointer(1),
				Comment:   fastly.ToPointer("a"),
				ServiceID: fastly.ToPointer("b"),
				CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
				UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-04T04:05:06Z"),
				DeletedAt: testutil.MustParseTimeRFC3339("2001-02-05T04:05:06Z"),
			},
			{
				Number:    fastly.ToPointer(2),
				Comment:   fastly.ToPointer("c"),
				ServiceID: fastly.ToPointer("d"),
				Active:    fastly.ToPointer(true),
				Deployed:  fastly.ToPointer(true),
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
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Deployed: true
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

var searchServiceVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

ID: 123
Name: Foo
Type: wasm
Customer ID: mycustomerid
Last edited (UTC): 2010-11-15 19:01
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2001-02-04 04:05
		Deleted (UTC): 2001-02-05 04:05
	Version 2/2
		Number: 2
		Comment: c
		Service ID: d
		Active: true
		Deployed: true
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2001-03-04 04:05
`) + "\n"

func updateServiceOK(_ *fastly.UpdateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ServiceID: fastly.ToPointer("12345"),
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
