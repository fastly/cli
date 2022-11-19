package domain_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestDomainCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("domain create --version 1"),
			WantError: "error reading service: no service ID found",
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
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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

func TestDomainValidate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("domain validate"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --token flag",
			Args:      args("domain validate --version 3"),
			WantError: fsterr.ErrNoToken.Inner.Error(),
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("domain validate --token 123 --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --name flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("domain validate --service-id 123 --token 123 --version 3"),
			WantError: "error parsing arguments: must provide --name flag",
		},
		{
			Name: "validate ValidateDomain API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ValidateDomainFn: func(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("domain validate --name foo.example.com --service-id 123 --token 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ValidateAllDomains API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ValidateAllDomainsFn: func(i *fastly.ValidateAllDomainsInput) ([]*fastly.DomainValidationResult, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("domain validate --all --service-id 123 --token 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ValidateDomain API success",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Args:       args("domain validate --name foo.example.com --service-id 123 --token 123 --version 3"),
			WantOutput: validateAPISuccess(3),
		},
		{
			Name: "validate ValidateAllDomains API success",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				ValidateAllDomainsFn: validateAllDomains,
			},
			Args:       args("domain validate --all --service-id 123 --token 123 --version 3"),
			WantOutput: validateAllAPISuccess(),
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Args:       args("domain validate --name foo.example.com --service-id 123 --token 123 --version 1"),
			WantOutput: validateAPISuccess(1),
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
		Name:           *i.Name,
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

var describeDomainOutput = "\n" + strings.TrimSpace(`
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

func validateDomain(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error) {
	return &fastly.DomainValidationResult{
		Metadata: fastly.DomainMetadata{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           i.Name,
		},
		CName: "foo",
		Valid: true,
	}, nil
}

func validateAllDomains(i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error) {
	return []*fastly.DomainValidationResult{
		{
			Metadata: fastly.DomainMetadata{
				ServiceID:      i.ServiceID,
				ServiceVersion: i.ServiceVersion,
				Name:           "foo.example.com",
			},
			CName: "foo",
			Valid: true,
		},
		{
			Metadata: fastly.DomainMetadata{
				ServiceID:      i.ServiceID,
				ServiceVersion: i.ServiceVersion,
				Name:           "bar.example.com",
			},
			CName: "bar",
			Valid: true,
		},
	}, nil
}

func validateAPISuccess(version int) string {
	return fmt.Sprintf(`
Service ID: 123
Service Version: %d

Name: foo.example.com
Valid: true
CNAME: foo`, version)
}

func validateAllAPISuccess() string {
	return `
Service ID: 123
Service Version: 3

Name: foo.example.com
Valid: true
CNAME: foo

Name: bar.example.com
Valid: true
CNAME: bar`
}
