package privatekey_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	mockFieldValue        = "example"
	mockKeyLength         = 123
	mockResponseID        = "123"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestTLSCustomPrivateKeyCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --key flag",
			Args:      args("tls-custom private-key create --name example"),
			WantError: "required flag --key not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      args("tls-custom private-key create --key example"),
			WantError: "required flag --name not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreatePrivateKeyFn: func(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom private-key create --key example --name example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreatePrivateKeyFn: func(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
					return &fastly.PrivateKey{
						ID:   mockResponseID,
						Name: i.Name,
					}, nil
				},
			},
			Args:       args("tls-custom private-key create --key example --name example"),
			WantOutput: "Created TLS Private Key 'example'",
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

func TestTLSCustomPrivateKeyDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom private-key delete"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeletePrivateKeyFn: func(_ *fastly.DeletePrivateKeyInput) error {
					return testutil.Err
				},
			},
			Args:      args("tls-custom private-key delete --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeletePrivateKeyFn: func(_ *fastly.DeletePrivateKeyInput) error {
					return nil
				},
			},
			Args:       args("tls-custom private-key delete --id example"),
			WantOutput: "Deleted TLS Private Key 'example'",
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

func TestTLSCustomPrivateKeyDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom private-key describe"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetPrivateKeyFn: func(_ *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom private-key describe --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetPrivateKeyFn: func(_ *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error) {
					t := testutil.Date
					return &fastly.PrivateKey{
						ID:            mockResponseID,
						Name:          mockFieldValue,
						KeyLength:     mockKeyLength,
						KeyType:       mockFieldValue,
						PublicKeySHA1: mockFieldValue,
						CreatedAt:     &t,
					}, nil
				},
			},
			Args:       args("tls-custom private-key describe --id example"),
			WantOutput: "\nID: " + mockResponseID + "\nName: example\nKey Length: 123\nKey Type: example\nPublic Key SHA1: example\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: false\n",
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

func TestTLSCustomPrivateKeyList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListPrivateKeysFn: func(_ *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom private-key list"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListPrivateKeysFn: func(_ *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error) {
					t := testutil.Date
					return []*fastly.PrivateKey{
						{
							ID:            mockResponseID,
							Name:          mockFieldValue,
							KeyLength:     mockKeyLength,
							KeyType:       mockFieldValue,
							PublicKeySHA1: mockFieldValue,
							CreatedAt:     &t,
						},
					}, nil
				},
			},
			Args:       args("tls-custom private-key list --verbose"),
			WantOutput: "\nID: " + mockResponseID + "\nName: example\nKey Length: 123\nKey Type: example\nPublic Key SHA1: example\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: false\n",
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
