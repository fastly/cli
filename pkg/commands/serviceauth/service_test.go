package serviceauth_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestServiceAuthCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth create"),
			wantError: "error parsing arguments: required flag --user-id not provided",
		},
		{
			args:      args("service-auth create --user-id 123 --service-id 123"),
			api:       mock.API{CreateServiceAuthorizationFn: createServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth create --user-id 123 --service-id 123"),
			api:        mock.API{CreateServiceAuthorizationFn: createServiceAuthOK},
			wantOutput: "Created service authorization 12345",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth list --verbose --json"),
			wantError: "invalid flag combination, --verbose and --json",
		},
		{
			args:      args("service-auth list"),
			api:       mock.API{ListServiceAuthorizationsFn: listServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth list"),
			api:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: "AUTH ID  USER ID  SERVICE ID  PERMISSION\n123      456      789         read_only\n",
		},
		{
			args:       args("service-auth list --json"),
			api:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: "{\"Info\":{\"links\":{},\"meta\":{}},\"Items\":[{\"CreatedAt\":null,\"DeletedAt\":null,\"ID\":\"123\",\"Permission\":\"read_only\",\"Service\":{\"ID\":\"789\"},\"UpdatedAt\":null,\"User\":{\"ID\":\"456\"}}]}",
		},
		{
			args:       args("service-auth list --verbose"),
			api:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: "Fastly API token not provided\nFastly API endpoint: https://api.fastly.com\n\nAuth ID: 123\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth describe"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth describe --id 123 --verbose --json"),
			wantError: "invalid flag combination, --verbose and --json",
		},
		{
			args:      args("service-auth describe --id 123"),
			api:       mock.API{GetServiceAuthorizationFn: describeServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth describe --id 123"),
			api:        mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			wantOutput: "Auth ID: 12345\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
		},
		{
			args:       args("service-auth describe --id 123 --json"),
			api:        mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			wantOutput: "{\"CreatedAt\":null,\"DeletedAt\":null,\"ID\":\"12345\",\"Permission\":\"read_only\",\"Service\":{\"ID\":\"789\"},\"UpdatedAt\":null,\"User\":{\"ID\":\"456\"}}",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth update --permission full"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth update --id 123"),
			wantError: "error parsing arguments: required flag --permission not provided",
		},
		{
			args:      args("service-auth update --id 123 --permission full"),
			api:       mock.API{UpdateServiceAuthorizationFn: updateServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth update --id 123 --permission full"),
			api:        mock.API{UpdateServiceAuthorizationFn: updateServiceAuthOK},
			wantOutput: "Updated service authorization 123",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth delete"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth delete --id 123"),
			api:       mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth delete --id 123"),
			api:        mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthOK},
			wantOutput: "Deleted service authorization 123",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createServiceAuthError(*fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func createServiceAuthOK(i *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func listServiceAuthError(*fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return nil, errTest
}

func listServiceAuthOK(i *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return &fastly.ServiceAuthorizations{
		Items: []*fastly.ServiceAuthorization{
			{
				ID: "123",
				User: &fastly.SAUser{
					ID: "456",
				},
				Service: &fastly.SAService{
					ID: "789",
				},
				Permission: "read_only",
			},
		},
	}, nil
}

func describeServiceAuthError(*fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func describeServiceAuthOK(i *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
		User: &fastly.SAUser{
			ID: "456",
		},
		Service: &fastly.SAService{
			ID: "789",
		},
		Permission: "read_only",
	}, nil
}

func updateServiceAuthError(*fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func updateServiceAuthOK(i *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func deleteServiceAuthError(*fastly.DeleteServiceAuthorizationInput) error {
	return errTest
}

func deleteServiceAuthOK(i *fastly.DeleteServiceAuthorizationInput) error {
	return nil
}
