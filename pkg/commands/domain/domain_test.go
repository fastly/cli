package domain_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
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
			Name:      "validate missing --service-id flag",
			Args:      args("domain validate --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --name flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("domain validate --service-id 123 --version 3"),
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
			Args:      args("domain validate --name foo.example.com --service-id 123 --version 3"),
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
			Args:      args("domain validate --all --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ValidateDomain API success",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Args:       args("domain validate --name foo.example.com --service-id 123 --version 3"),
			WantOutput: validateAPISuccess(3),
		},
		{
			Name: "validate ValidateAllDomains API success",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				ValidateAllDomainsFn: validateAllDomains,
			},
			Args:       args("domain validate --all --service-id 123 --version 3"),
			WantOutput: validateAllAPISuccess(),
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Args:       args("domain validate --name foo.example.com --service-id 123 --version 1"),
			WantOutput: validateAPISuccess(1),
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createDomainError(_ *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func listDomainsOK(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("www.test.com"),
			Comment:        fastly.ToPointer("test"),
		},
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("www.example.com"),
			Comment:        fastly.ToPointer("example"),
		},
	}, nil
}

func listDomainsError(_ *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, errTest
}

var listDomainsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME             COMMENT
123      1        www.test.com     test
123      1        www.example.com  example
`) + "\n"

var listDomainsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

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
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		Comment:        fastly.ToPointer("test"),
	}, nil
}

func getDomainError(_ *fastly.GetDomainInput) (*fastly.Domain, error) {
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
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.NewName,
	}, nil
}

func updateDomainError(_ *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func deleteDomainOK(_ *fastly.DeleteDomainInput) error {
	return nil
}

func deleteDomainError(_ *fastly.DeleteDomainInput) error {
	return errTest
}

func validateDomain(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error) {
	return &fastly.DomainValidationResult{
		Metadata: &fastly.DomainMetadata{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer(i.Name),
		},
		CName: fastly.ToPointer("foo"),
		Valid: fastly.ToPointer(true),
	}, nil
}

func validateAllDomains(i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error) {
	return []*fastly.DomainValidationResult{
		{
			Metadata: &fastly.DomainMetadata{
				ServiceID:      fastly.ToPointer(i.ServiceID),
				ServiceVersion: fastly.ToPointer(i.ServiceVersion),
				Name:           fastly.ToPointer("foo.example.com"),
			},
			CName: fastly.ToPointer("foo"),
			Valid: fastly.ToPointer(true),
		},
		{
			Metadata: &fastly.DomainMetadata{
				ServiceID:      fastly.ToPointer(i.ServiceID),
				ServiceVersion: fastly.ToPointer(i.ServiceVersion),
				Name:           fastly.ToPointer("bar.example.com"),
			},
			CName: fastly.ToPointer("bar"),
			Valid: fastly.ToPointer(true),
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
