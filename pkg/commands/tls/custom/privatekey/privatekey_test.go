package privatekey_test

import (
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/custom"
	sub "github.com/fastly/cli/pkg/commands/tls/custom/privatekey"
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
	var content string
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --key and --key-path flags",
			Args:      "--name example",
			WantError: "neither --key-path or --key provided, one must be provided",
		},
		{
			Name:      "validate using both --key and --key-path flags",
			Args:      "--name example --key example --key-path foobar",
			WantError: "--key-path and --key provided, only one can be specified",
		},
		{
			Name:      "validate missing --name flag",
			Args:      "--key example",
			WantError: "required flag --name not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreatePrivateKeyFn: func(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
					content = i.Key
					return nil, testutil.Err
				},
			},
			Args:            "--key example --name example",
			WantError:       testutil.Err.Error(),
			PathContentFlag: &testutil.PathContentFlag{Flag: "key-path", Fixture: "testkey.pem", Content: func() string { return content }},
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreatePrivateKeyFn: func(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
					content = i.Key
					return &fastly.PrivateKey{
						ID:   mockResponseID,
						Name: i.Name,
					}, nil
				},
			},
			Args:            "--key example --name example",
			WantOutput:      "Created TLS Private Key 'example'",
			PathContentFlag: &testutil.PathContentFlag{Flag: "key-path", Fixture: "testkey.pem", Content: func() string { return content }},
		},
		{
			Name: "validate custom key is submitted",
			API: mock.API{
				CreatePrivateKeyFn: func(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
					content = i.Key
					return &fastly.PrivateKey{
						ID:   mockResponseID,
						Name: i.Name,
					}, nil
				},
			},
			Args:            "--name example --key-path ./testdata/testkey.pem",
			WantOutput:      "Created TLS Private Key 'example'",
			PathContentFlag: &testutil.PathContentFlag{Flag: "key-path", Fixture: "testkey.pem", Content: func() string { return content }},
		},
		{
			Name:      "validate invalid --key-path arg",
			Args:      "--name example --key-path ............",
			WantError: "error reading key-path",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestTLSCustomPrivateKeyDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeletePrivateKeyFn: func(_ *fastly.DeletePrivateKeyInput) error {
					return testutil.Err
				},
			},
			Args:      "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeletePrivateKeyFn: func(_ *fastly.DeletePrivateKeyInput) error {
					return nil
				},
			},
			Args:       "--id example",
			WantOutput: "Deleted TLS Private Key 'example'",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestTLSCustomPrivateKeyDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetPrivateKeyFn: func(_ *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id example",
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
			Args:       "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nName: example\nKey Length: 123\nKey Type: example\nPublic Key SHA1: example\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: false\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestTLSCustomPrivateKeyList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListPrivateKeysFn: func(_ *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error) {
					return nil, testutil.Err
				},
			},
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
			Args:       "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nName: example\nKey Length: 123\nKey Type: example\nPublic Key SHA1: example\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: false\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}
