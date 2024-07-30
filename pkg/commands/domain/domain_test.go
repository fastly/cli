package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/domain"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const ()

func TestDomainCreate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateDomainFn: createDomainOK,
			},
			WantOutput: "Created domain www.test.com (service 123 version 4)",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateDomainFn: createDomainError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDomainList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsShortOutput,
		},
		{
			Arg: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Arg: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Arg: "--verbose --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Arg: "-v --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOK,
			},
			WantOutput: listDomainsVerboseOutput,
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestDomainDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDomainFn:    getDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDomainFn:    getDomainOK,
			},
			WantOutput: describeDomainOutput,
		},
	}
	testutil.RunScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestDomainUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123 --version 1 --new-name www.test.com --comment ",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainOK,
			},
			WantError: "error parsing arguments: must provide either --new-name or --comment to update domain",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateDomainFn: updateDomainOK,
			},
			WantOutput: "Updated domain www.example.com (service 123 version 4)",
		},
	}
	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func TestDomainDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteDomainFn: deleteDomainError,
			},
			WantError: errTest.Error(),
		},
		{
			Arg: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteDomainFn: deleteDomainOK,
			},
			WantOutput: "Deleted domain www.test.com (service 123 version 4)",
		},
	}
	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestDomainValidate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --name flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--service-id 123 --version 3",
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
			Arg:       "--name foo.example.com --service-id 123 --version 3",
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
			Arg:       "--all --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ValidateDomain API success",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Arg:        "--name foo.example.com --service-id 123 --version 3",
			WantOutput: validateAPISuccess(3),
		},
		{
			Name: "validate ValidateAllDomains API success",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				ValidateAllDomainsFn: validateAllDomains,
			},
			Arg:        "--all --service-id 123 --version 3",
			WantOutput: validateAllAPISuccess(),
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ValidateDomainFn: validateDomain,
			},
			Arg:        "--name foo.example.com --service-id 123 --version 1",
			WantOutput: validateAPISuccess(1),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "validate"}, scenarios)
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
