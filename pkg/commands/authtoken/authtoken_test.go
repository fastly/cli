package authtoken_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --username flag",
			Args:      args("auth-token create --password secure"),
			WantError: "error parsing arguments: required flag --username not provided",
		},
		{
			Name:      "validate missing --password flag",
			Args:      args("auth-token create --username foo@example.com"),
			WantError: "error parsing arguments: required flag --password not provided",
		},
		{
			Name: "validate CreateToken API error",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("auth-token create --username foo@example.com --password secure"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateToken API success",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return &fastly.Token{
						ExpiresAt: &testutil.Date,
						ID:        "123",
						Name:      "Example",
						Scope:     "foobar",
					}, nil
				},
			},
			Args:       args("auth-token create --username foo@example.com --password secure"),
			WantOutput: "Created token 'Example' (id: 123, scope: foobar, expires: 2021-06-15 23:00:00 +0000 UTC)",
		},
		{
			Name: "validate CreateToken API success with all flags",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return &fastly.Token{
						ExpiresAt: i.ExpiresAt,
						ID:        "123",
						Name:      i.Name,
						Scope:     i.Scope,
					}, nil
				},
			},
			Args:       args("auth-token create --username foo@example.com --password secure --expires 2021-09-15T23:00:00Z --name Testing --scope purge_all,global:read --services a,b,c --token 123"),
			WantOutput: "Created token 'Testing' (id: 123, scope: purge_all global:read, expires: 2021-09-15 23:00:00 +0000 UTC)",
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

func TestDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("auth-token delete"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name:      "validate missing optional flags",
			Args:      args("auth-token delete --token 123"),
			WantError: "error parsing arguments: must provide either the --current, --file or --id flag",
		},
		{
			Name: "validate DeleteTokenSelf API error with --current",
			API: mock.API{
				DeleteTokenSelfFn: func() error {
					return testutil.Err
				},
			},
			Args:      args("auth-token delete --current --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate BatchDeleteTokens API error with --file",
			API: mock.API{
				BatchDeleteTokensFn: func(i *fastly.BatchDeleteTokensInput) error {
					return testutil.Err
				},
			},
			Args:      args("auth-token delete --file ./testdata/tokens --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteToken API error with --id",
			API: mock.API{
				DeleteTokenFn: func(i *fastly.DeleteTokenInput) error {
					return testutil.Err
				},
			},
			Args:      args("auth-token delete --id 123 --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteTokenSelf API success with --current",
			API: mock.API{
				DeleteTokenSelfFn: func() error {
					return nil
				},
			},
			Args:       args("auth-token delete --current --token 123"),
			WantOutput: "Deleted current token",
		},
		{
			Name: "validate BatchDeleteTokens API success with --file",
			API: mock.API{
				BatchDeleteTokensFn: func(i *fastly.BatchDeleteTokensInput) error {
					return nil
				},
			},
			Args:       args("auth-token delete --file ./testdata/tokens --token 123"),
			WantOutput: "Deleted tokens",
		},
		{
			Name: "validate BatchDeleteTokens API success with --file and --verbose",
			API: mock.API{
				BatchDeleteTokensFn: func(i *fastly.BatchDeleteTokensInput) error {
					return nil
				},
			},
			Args:       args("auth-token delete --file ./testdata/tokens --token 123 --verbose"),
			WantOutput: fileTokensOutput(),
		},
		{
			Name: "validate DeleteToken API success with --id",
			API: mock.API{
				DeleteTokenFn: func(i *fastly.DeleteTokenInput) error {
					return nil
				},
			},
			Args:       args("auth-token delete --id 123 --token 123"),
			WantOutput: "Deleted token '123'",
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

func TestDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("auth-token describe"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name: "validate GetTokenSelf API error",
			API: mock.API{
				GetTokenSelfFn: func() (*fastly.Token, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("auth-token describe --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetTokenSelf API success",
			API: mock.API{
				GetTokenSelfFn: getToken,
			},
			Args:       args("auth-token describe --token 123"),
			WantOutput: describeTokenOutput(),
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

func TestList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("auth-token list"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name: "validate ListTokens API error",
			API: mock.API{
				ListTokensFn: func() ([]*fastly.Token, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("auth-token list --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListCustomerTokens API error",
			API: mock.API{
				ListCustomerTokensFn: func(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("auth-token list --customer-id 123 --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListTokens API success",
			API: mock.API{
				ListTokensFn: listTokens,
			},
			Args:       args("auth-token list --token 123"),
			WantOutput: listTokenOutputSummary(),
		},
		{
			Name: "validate ListCustomerTokens API success",
			API: mock.API{
				ListCustomerTokensFn: listCustomerTokens,
			},
			Args:       args("auth-token list --customer-id 123 --token 123"),
			WantOutput: listTokenOutputSummary(),
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListTokensFn: listTokens,
			},
			Args:       args("auth-token list --token 123 --verbose"),
			WantOutput: listTokenOutputVerbose(),
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

func getToken() (*fastly.Token, error) {
	t := testutil.Date

	return &fastly.Token{
		ID:          "123",
		Name:        "Foo",
		UserID:      "456",
		Services:    []string{"a", "b"},
		AccessToken: "xyz",
		Scope:       fastly.TokenScope(fmt.Sprintf("%s %s", fastly.PurgeAllScope, fastly.GlobalReadScope)),
		IP:          "127.0.0.1",
		CreatedAt:   &t,
		ExpiresAt:   &t,
		LastUsedAt:  &t,
	}, nil
}

func listTokens() ([]*fastly.Token, error) {
	t := testutil.Date
	token, _ := getToken()
	vs := []*fastly.Token{
		token,
		{
			ID:          "456",
			Name:        "Bar",
			UserID:      "789",
			Services:    []string{"a", "b"},
			AccessToken: "999",
			Scope:       fastly.GlobalScope,
			IP:          "127.0.0.2",
			CreatedAt:   &t,
			ExpiresAt:   &t,
			LastUsedAt:  &t,
		},
	}
	return vs, nil
}

func listCustomerTokens(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
	return listTokens()
}

func fileTokensOutput() string {
	return `Deleted tokens

TOKEN ID
abc
def
xyz`
}

func describeTokenOutput() string {
	return `
ID: 123
Name: Foo
User ID: 456
Services: a, b
Access Token: xyz
Scope: purge_all global:read
IP: 127.0.0.1

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC`
}

func listTokenOutputVerbose() string {
	return `Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com

ID: 123
Name: Foo
User ID: 456
Services: a, b
Access Token: xyz
Scope: purge_all global:read
IP: 127.0.0.1

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC

ID: 456
Name: Bar
User ID: 789
Services: a, b
Access Token: 999
Scope: global
IP: 127.0.0.2

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC`
}

func listTokenOutputSummary() string {
	return `TOKEN ID  USER ID  SCOPE                  SERVICES
123       456      purge_all global:read  [a b]
456       789      global                 [a b]`
}
