// NOTE: We always pass the --token flag as this allows us to side-step the
// browser based authentication flow. This is because if a token is explicitly
// provided, then we respect the user knows what they're doing.
package service_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestServiceCreate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service create --token 123"),
			api:       mock.API{CreateServiceFn: createServiceOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       args("service create --name Foo --token 123"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create -n=Foo --token 123"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create --name Foo --type wasm --token 123"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create --name Foo --type wasm --comment Hello --token 123"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:       args("service create -n Foo --comment Hello --token 123"),
			api:        mock.API{CreateServiceFn: createServiceOK},
			wantOutput: "Created service 12345",
		},
		{
			args:      args("service create -n Foo --token 123"),
			api:       mock.API{CreateServiceFn: createServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

type mockServicesPaginator struct {
	count         int
	maxPages      int
	numOfPages    int
	requestedPage int
	returnErr     bool
}

func (p *mockServicesPaginator) HasNext() bool {
	if p.count > p.maxPages {
		return false
	}
	p.count++
	return true
}

func (p mockServicesPaginator) Remaining() int {
	return 1
}

func (p *mockServicesPaginator) GetNext() (ss []*fastly.Service, err error) {
	if p.returnErr {
		err = testutil.Err
	}
	pageOne := fastly.Service{
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
	}
	pageTwo := fastly.Service{
		ID:            "456",
		Name:          "Bar",
		Type:          "wasm",
		CustomerID:    "mycustomerid",
		ActiveVersion: 1,
		UpdatedAt:     testutil.MustParseTimeRFC3339("2015-03-14T12:59:59Z"),
	}
	pageThree := fastly.Service{
		ID:            "789",
		Name:          "Baz",
		Type:          "vcl",
		CustomerID:    "mycustomerid",
		ActiveVersion: 1,
	}
	if p.count == 1 {
		ss = append(ss, &pageOne)
	}
	if p.count == 2 {
		ss = append(ss, &pageTwo)
	}
	if p.count == 3 {
		ss = append(ss, &pageThree)
	}
	if p.requestedPage > 0 && p.numOfPages == 1 {
		p.count = p.maxPages + 1 // forces only one result to be displayed
	}
	return ss, err
}

func TestServiceList(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			api: mock.API{
				NewListServicesPaginatorFn: func(i *fastly.ListServicesInput) fastly.PaginatorServices {
					return &mockServicesPaginator{returnErr: true}
				},
			},
			args:      args("service list --token 123"),
			wantError: testutil.Err.Error(),
		},
		// NOTE: Our mock paginator defines three services, and so even when setting
		// --per-page 1 we expect the final output to display both items.
		{
			api: mock.API{
				NewListServicesPaginatorFn: func(i *fastly.ListServicesInput) fastly.PaginatorServices {
					return &mockServicesPaginator{numOfPages: i.PerPage, maxPages: 3}
				},
			},
			args:       args("service list --per-page 1 --token 123"),
			wantOutput: listServicesShortOutput,
		},
		// In the following test, we set --page 1 and as there's only one record
		// displayed per page we expect only the first record to be displayed.
		{
			api: mock.API{
				NewListServicesPaginatorFn: func(i *fastly.ListServicesInput) fastly.PaginatorServices {
					return &mockServicesPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 3}
				},
			},
			args:       args("service list --page 1 --per-page 1 --token 123"),
			wantOutput: listServicesShortOutputPageOne,
		},
		// In the following test, we set --page 2 and as there's only one record
		// displayed per page we expect only the second record to be displayed.
		{
			api: mock.API{
				NewListServicesPaginatorFn: func(i *fastly.ListServicesInput) fastly.PaginatorServices {
					return &mockServicesPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 3}
				},
			},
			args:       args("service list --page 2 --per-page 1 --token 123"),
			wantOutput: listServicesShortOutputPageTwo,
		},
		{
			api: mock.API{
				NewListServicesPaginatorFn: func(i *fastly.ListServicesInput) fastly.PaginatorServices {
					return &mockServicesPaginator{maxPages: 3}
				},
			},
			args:       args("service list --verbose --token 123"),
			wantOutput: listServicesVerboseOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestServiceDescribe(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service describe --token 123"),
			api:       mock.API{GetServiceDetailsFn: describeServiceOK},
			wantError: "error reading service: no service ID found",
		},
		{
			args:       args("service describe --service-id 123 --token 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceShortOutput,
		},
		{
			args:       args("service describe --service-id 123 --verbose --token 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("service describe --service-id 123 -v --token 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("service --verbose describe --service-id 123 --token 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:       args("-v service describe --service-id 123 --token 123"),
			api:        mock.API{GetServiceDetailsFn: describeServiceOK},
			wantOutput: describeServiceVerboseOutput,
		},
		{
			args:      args("service describe --service-id 123 --token 123"),
			api:       mock.API{GetServiceDetailsFn: describeServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestServiceSearch(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service search --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       args("service search --name Foo --token 123"),
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceShortOutput,
		},
		{
			args:       args("service search --name Foo -v --token 123"),
			api:        mock.API{SearchServiceFn: searchServiceOK},
			wantOutput: searchServiceVerboseOutput,
		},
		{
			args:      args("service search --name --token 123"),
			api:       mock.API{SearchServiceFn: searchServiceOK},
			wantError: "error parsing arguments: expected argument for flag '--name'",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("service update --token 123"),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateServiceFn: updateServiceOK,
			},
			wantError: "error reading service: no service ID found",
		},
		{
			args:      args("service update --service-id 12345 --token 123"),
			api:       mock.API{UpdateServiceFn: updateServiceOK},
			wantError: "error parsing arguments: must provide either --name or --comment to update service",
		},
		{
			args:       args("service update --service-id 12345 --name Foo --token 123"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 -n=Foo --token 123"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 --name Foo --token 123"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 --name Foo --comment Hello --token 123"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:       args("service update --service-id 12345 -n Foo --comment Hello --token 123"),
			api:        mock.API{UpdateServiceFn: updateServiceOK},
			wantOutput: "Updated service 12345",
		},
		{
			args:      args("service update --service-id 12345 -n Foo --token 123"),
			api:       mock.API{UpdateServiceFn: updateServiceError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceDelete(t *testing.T) {
	args := testutil.Args
	nonEmptyServiceID := regexp.MustCompile(`service_id = "[^"]+"`)

	for _, testcase := range []struct {
		args                 []string
		api                  mock.API
		manifest             string
		wantError            string
		wantOutput           string
		expectEmptyServiceID bool
	}{
		{
			args:      args("service delete --token 123"),
			api:       mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:  "fastly-no-serviceid.toml",
			wantError: "error reading service: no service ID found",
		},
		{
			args:                 args("service delete --token 123"),
			api:                  mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:             "fastly-valid.toml",
			wantOutput:           "Deleted service ID 123",
			expectEmptyServiceID: true,
		},
		{
			args:       args("service delete --service-id 001 --token 123"),
			api:        mock.API{DeleteServiceFn: deleteServiceOK},
			wantOutput: "Deleted service ID 001",
		},
		{
			args:                 args("service delete --service-id 001 --token 123"),
			api:                  mock.API{DeleteServiceFn: deleteServiceOK},
			manifest:             "fastly-valid.toml",
			wantOutput:           "Deleted service ID 001",
			expectEmptyServiceID: false,
		},
		{
			args:      args("service delete --service-id 001 --token 123"),
			api:       mock.API{DeleteServiceFn: deleteServiceError},
			manifest:  "fastly-valid.toml",
			wantError: errTest.Error(),
		},
	} {
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
			defer os.Chdir(pwd)

			var stdout bytes.Buffer
			runOpts := testutil.NewRunOpts(testcase.args, &stdout)
			runOpts.ClientFactory = mock.ClientFactory(testcase.api)
			runErr := app.Run(runOpts)
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
		ID:      "12345",
		Name:    i.Name,
		Type:    i.Type,
		Comment: i.Comment,
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

var listServicesShortOutput = strings.TrimSpace(`
NAME  ID   TYPE  ACTIVE VERSION  LAST EDITED (UTC)
Foo   123  wasm  2               2010-11-15 19:01
Bar   456  wasm  1               2015-03-14 12:59
Baz   789  vcl   1               n/a
`) + "\n"

var listServicesShortOutputPageOne = strings.TrimSpace(`
NAME  ID   TYPE  ACTIVE VERSION  LAST EDITED (UTC)
Foo   123  wasm  2               2010-11-15 19:01
`) + "\n"

var listServicesShortOutputPageTwo = strings.TrimSpace(`
NAME  ID   TYPE  ACTIVE VERSION  LAST EDITED (UTC)
Bar   456  wasm  1               2015-03-14 12:59
`) + "\n"

var listServicesVerboseOutput = strings.TrimSpace(`
Fastly API token provided via --token
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
Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com
Service ID (via --service-id): 123

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
Fastly API token provided via --token
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
		ID: "12345",
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
