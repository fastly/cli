package serviceauth_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
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

var errTest = errors.New("fixture error")

func createServiceAuthOK(i *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func createServiceAuthError(*fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}
