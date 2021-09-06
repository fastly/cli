package domain_test

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

func TestDomainCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("domain create --version 1 --service-id 123"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("domain create --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateDomainFn: createDomainOK,
			},
			WantOutput: "Created domain www.test.com (service 123 version 4)",
		},
		{
			Args: args("domain create --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateDomainFn: createDomainError,
			},
			WantError: errTest.Error(),
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestDomainList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args: args("domain list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsShortOutput,
		},
		{
			Args: args("domain list --service-id 123 --version 1 --verbose"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Args: args("domain list --service-id 123 --version 1 -v"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Args: args("domain --verbose list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Args: args("-v domain list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Args: args("domain list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsError,
			},
			WantError: errTest.Error(),
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDomainDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("domain describe --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("domain describe --service-id 123 --version 1 --name www.test.com"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDomainFn:    getDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("domain describe --service-id 123 --version 1 --name www.test.com"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDomainFn:    getDomainOK,
			},
			WantOutput: describeDomainOutput,
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDomainUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("domain update --service-id 123 --version 1 --new-name www.test.com --comment "),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("domain update --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainOK,
			},
			WantError: "error parsing arguments: must provide either --new-name or --comment to update domain",
		},
		{
			Args: args("domain update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("domain update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainOK,
			},
			WantOutput: "Updated domain www.example.com (service 123 version 4)",
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestDomainDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("domain delete --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("domain delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteDomainFn: deleteDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("domain delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteDomainFn: deleteDomainOK,
			},
			WantOutput: "Deleted domain www.test.com (service 123 version 4)",
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        i.Comment,
	}, nil
}

func createDomainError(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func listDomainsOK(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "www.test.com",
			Comment:        "test",
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "www.example.com",
			Comment:        "example",
		},
	}, nil
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, errTest
}

var listDomainsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME             COMMENT
123      1        www.test.com     test
123      1        www.example.com  example
`) + "\n"

var listDomainsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID (via --service-id): 123

Version: 1
	Domain 1/2
		Name: www.test.com
		Comment: test
	Domain 2/2
		Name: www.example.com
		Comment: example
`) + "\n\n"

func getDomainOK(i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        "test",
	}, nil
}

func getDomainError(i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

var describeDomainOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: www.test.com
Comment: test
`) + "\n"

func updateDomainOK(i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.NewName,
	}, nil
}

func updateDomainError(i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func deleteDomainOK(i *fastly.DeleteDomainInput) error {
	return nil
}

func deleteDomainError(i *fastly.DeleteDomainInput) error {
	return errTest
}
