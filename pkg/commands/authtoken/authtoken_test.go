package authtoken_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAuthTokenCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --password flag",
			Args:      args("auth-token create"),
			WantError: "error parsing arguments: required flag --password not provided",
		},
		{
			Name: "validate CreateToken API error",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("auth-token create --password secure --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateToken API success with no flags",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return &fastly.Token{
						ExpiresAt:   &testutil.Date,
						TokenID:     fastly.ToPointer("123"),
						Name:        fastly.ToPointer("Example"),
						Scope:       fastly.ToPointer(fastly.TokenScope("foobar")),
						AccessToken: fastly.ToPointer("123abc"),
					}, nil
				},
			},
			Args:       args("auth-token create --password secure --token 123"),
			WantOutput: "Created token '123abc' (name: Example, id: 123, scope: foobar, expires: 2021-06-15 23:00:00 +0000 UTC)",
		},
		{
			Name: "validate CreateToken API success with all flags",
			API: mock.API{
				CreateTokenFn: func(i *fastly.CreateTokenInput) (*fastly.Token, error) {
					return &fastly.Token{
						ExpiresAt:   i.ExpiresAt,
						TokenID:     fastly.ToPointer("123"),
						Name:        i.Name,
						Scope:       i.Scope,
						AccessToken: fastly.ToPointer("123abc"),
					}, nil
				},
			},
			Args:       args("auth-token create --expires 2021-09-15T23:00:00Z --name Testing --password secure --scope purge_all --scope global:read --services a,b,c --token 123"),
			WantOutput: "Created token '123abc' (name: Testing, id: 123, scope: purge_all global:read, expires: 2021-09-15 23:00:00 +0000 UTC)",
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

func TestAuthTokenDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
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

func TestAuthTokenDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
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

func TestAuthTokenList(t *testing.T) {
	args := testutil.Args
	type ts struct {
		testutil.TestScenario
		SetEnv bool
	}
	scenarios := []ts{
		{
			TestScenario: testutil.TestScenario{
				Name: "validate ListTokens API error",
				API: mock.API{
					ListTokensFn: func(_ *fastly.ListTokensInput) ([]*fastly.Token, error) {
						return nil, testutil.Err
					},
				},
				Args:      args("auth-token list"),
				WantError: testutil.Err.Error(),
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate ListCustomerTokens API error",
				API: mock.API{
					ListCustomerTokensFn: func(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
						return nil, testutil.Err
					},
				},
				Args:      args("auth-token list --customer-id 123"),
				WantError: testutil.Err.Error(),
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate ListTokens API success",
				API: mock.API{
					ListTokensFn: listTokens,
				},
				Args:       args("auth-token list"),
				WantOutput: listTokenOutputSummary(false),
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate ListCustomerTokens API success",
				API: mock.API{
					ListCustomerTokensFn: listCustomerTokens,
				},
				Args:       args("auth-token list --customer-id 123"),
				WantOutput: listTokenOutputSummary(false),
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate ListCustomerTokens API success with env var",
				API: mock.API{
					ListCustomerTokensFn: listCustomerTokens,
				},
				Args:       args("auth-token list"),
				WantOutput: listTokenOutputSummary(true),
			},
			SetEnv: true,
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate --verbose flag",
				API: mock.API{
					ListTokensFn: listTokens,
				},
				Args:       args("auth-token list --verbose"),
				WantOutput: listTokenOutputVerbose(),
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			if testcase.SetEnv {
				if err := os.Setenv("FASTLY_CUSTOMER_ID", "123"); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Unsetenv("FASTLY_CUSTOMER_ID"); err != nil {
						t.Fatal(err)
					}
				}()
			}
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

func getToken() (*fastly.Token, error) {
	t := testutil.Date

	return &fastly.Token{
		TokenID:    fastly.ToPointer("123"),
		Name:       fastly.ToPointer("Foo"),
		UserID:     fastly.ToPointer("456"),
		Services:   []string{"a", "b"},
		Scope:      fastly.ToPointer(fastly.TokenScope(fmt.Sprintf("%s %s", fastly.PurgeAllScope, fastly.GlobalReadScope))),
		IP:         fastly.ToPointer("127.0.0.1"),
		CreatedAt:  &t,
		ExpiresAt:  &t,
		LastUsedAt: &t,
	}, nil
}

func listTokens(_ *fastly.ListTokensInput) ([]*fastly.Token, error) {
	t := testutil.Date
	token, _ := getToken()
	vs := []*fastly.Token{
		token,
		{
			TokenID:    fastly.ToPointer("456"),
			Name:       fastly.ToPointer("Bar"),
			UserID:     fastly.ToPointer("789"),
			Services:   []string{"a", "b"},
			Scope:      fastly.ToPointer(fastly.GlobalScope),
			IP:         fastly.ToPointer("127.0.0.2"),
			CreatedAt:  &t,
			ExpiresAt:  &t,
			LastUsedAt: &t,
		},
	}
	return vs, nil
}

func listCustomerTokens(_ *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
	return listTokens(nil)
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
Scope: purge_all global:read
IP: 127.0.0.1

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC`
}

func listTokenOutputVerbose() string {
	return `Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)


ID: 123
Name: Foo
User ID: 456
Services: a, b
Scope: purge_all global:read
IP: 127.0.0.1

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC

ID: 456
Name: Bar
User ID: 789
Services: a, b
Scope: global
IP: 127.0.0.2

Created at: 2021-06-15 23:00:00 +0000 UTC
Last used at: 2021-06-15 23:00:00 +0000 UTC
Expires at: 2021-06-15 23:00:00 +0000 UTC

`
}

func listTokenOutputSummary(env bool) string {
	var msg string
	if env {
		msg = "INFO: Listing customer tokens for the FASTLY_CUSTOMER_ID environment variable\n\n"
	}
	return fmt.Sprintf(`%sNAME  TOKEN ID  USER ID  SCOPE                  SERVICES
Foo   123       456      purge_all global:read  a, b
Bar   456       789      global                 a, b`, msg)
}
