package secretstore_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/secretstore"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestCreateSecretCommand(t *testing.T) {
	const (
		storeID      = "store123"
		secretName   = "testsecret"
		secretDigest = "digest"
		secretValue  = "the secret"
	)

	tmpDir := t.TempDir()
	secretFile := path.Join(tmpDir, "secret-file")
	if err := os.WriteFile(secretFile, []byte(secretValue), 0x777); err != nil {
		t.Fatal(err)
	}
	doesNotExistFile := path.Join(tmpDir, "DOES-NOT-EXIST")

	scenarios := []struct {
		args           string
		stdin          string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "create --name test",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args:      "create --store-id abc123",
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s", storeID, secretName, doesNotExistFile),
			wantError: func() string {
				if runtime.GOOS == "windows" {
					return "The system cannot find the file specified"
				}
				return "no such file or directory"
			}(),
		},
		{
			args:      fmt.Sprintf("create --store-id %s --name %s --stdin", storeID, secretName),
			wantError: "unable to read from STDIN",
		},
		// Read from STDIN.
		{
			args:  fmt.Sprintf("create --store-id %s --name %s --stdin", storeID, secretName),
			stdin: secretValue,
			api: mock.API{
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if secret := string(i.Secret); secret != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", secret)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtSuccess("Created secret %s in store %s (digest %s)", secretName, storeID, hex.EncodeToString([]byte(secretDigest))),
		},
		// Read from file.
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s", storeID, secretName, secretFile),
			api: mock.API{
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if secret := string(i.Secret); secret != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", secret)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtSuccess("Created secret %s in store %s (digest %s)", secretName, storeID, hex.EncodeToString([]byte(secretDigest))),
		},
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s --json", storeID, secretName, secretFile),
			api: mock.API{
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if secret := string(i.Secret); secret != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", secret)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: encodeJSON(&fastly.Secret{
				Name:   secretName,
				Digest: []byte(secretDigest),
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameSecret+" "+testcase.args), &stdout)
			if testcase.stdin != "" {
				var stdin bytes.Buffer
				stdin.WriteString(testcase.stdin)
				opts.Stdin = &stdin
			}

			f := testcase.api.CreateSecretFn
			var apiInvoked bool
			testcase.api.CreateSecretFn = func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestGetSecretCommand(t *testing.T) {
	const (
		storeID     = "testid"
		storeName   = "testname"
		storeDigest = "testdigest"
	)

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "get --store-id abc",
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      "get --name abc",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("get --store-id %s --name %s", "DOES-NOT-EXIST", storeName),
			api: mock.API{
				GetSecretFn: func(i *fastly.GetSecretInput) (*fastly.Secret, error) {
					if i.ID != storeID || i.Name != storeName {
						return nil, errors.New("invalid request")
					}
					return &fastly.Secret{
						Name:   storeName,
						Digest: []byte(storeDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("get --store-id %s --name %s", storeID, storeName),
			api: mock.API{
				GetSecretFn: func(i *fastly.GetSecretInput) (*fastly.Secret, error) {
					if i.ID != storeID || i.Name != storeName {
						return nil, errors.New("invalid request")
					}
					return &fastly.Secret{
						Name:   storeName,
						Digest: []byte(storeDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fmtSecret(&fastly.Secret{
				Name:   storeName,
				Digest: []byte(storeDigest),
			}),
		},
		{
			args: fmt.Sprintf("get --store-id %s --name %s --json", storeID, storeName),
			api: mock.API{
				GetSecretFn: func(i *fastly.GetSecretInput) (*fastly.Secret, error) {
					if i.ID != storeID || i.Name != storeName {
						return nil, errors.New("invalid request")
					}
					return &fastly.Secret{
						Name:   storeName,
						Digest: []byte(storeDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: encodeJSON(&fastly.Secret{
				Name:   storeName,
				Digest: []byte(storeDigest),
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameSecret+" "+testcase.args), &stdout)

			f := testcase.api.GetSecretFn
			var apiInvoked bool
			testcase.api.GetSecretFn = func(i *fastly.GetSecretInput) (*fastly.Secret, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API GetSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDeleteSecretCommand(t *testing.T) {
	const (
		storeID    = "test123"
		secretName = "testName"
	)

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "delete --name test",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args:      "delete --store-id test",
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: fmt.Sprintf("delete --store-id %s --name DOES-NOT-EXIST", storeID),
			api: mock.API{
				DeleteSecretFn: func(i *fastly.DeleteSecretInput) error {
					if i.ID != storeID || i.Name != secretName {
						return errors.New("not found")
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantError:      "not found",
		},
		{
			args: fmt.Sprintf("delete --store-id %s --name %s", storeID, secretName),
			api: mock.API{
				DeleteSecretFn: func(i *fastly.DeleteSecretInput) error {
					if i.ID != storeID || i.Name != secretName {
						return errors.New("not found")
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtSuccess("Deleted secret %s from store %s\n", secretName, storeID),
		},
		{
			args: fmt.Sprintf("delete --store-id %s --name %s --json", storeID, secretName),
			api: mock.API{
				DeleteSecretFn: func(i *fastly.DeleteSecretInput) error {
					if i.ID != storeID || i.Name != secretName {
						return errors.New("not found")
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtJSON(`{"name": %q, "store_id": %q,  "deleted": true}`, secretName, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameSecret+" "+testcase.args), &stdout)

			f := testcase.api.DeleteSecretFn
			var apiInvoked bool
			testcase.api.DeleteSecretFn = func(i *fastly.DeleteSecretInput) error {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestListSecretsCommand(t *testing.T) {
	const (
		secretName = "test123"
		storeID    = "store-id-123"
	)

	secrets := &fastly.Secrets{
		Meta: fastly.SecretStoreMeta{
			Limit:      123,
			NextCursor: "abc",
		},
		Data: []fastly.Secret{
			{Name: secretName, Digest: []byte(secretName)},
		},
	}

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "list",
			wantError: "required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("list --store-id %s", storeID),
			api: mock.API{
				ListSecretsFn: func(i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
					return secrets, errors.New("unknown error")
				},
			},
			wantAPIInvoked: true,
			wantError:      "unknown error",
		},
		{
			args: fmt.Sprintf("list --store-id %s", storeID),
			api: mock.API{
				ListSecretsFn: func(i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
					return secrets, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtSecrets(secrets),
		},
		{
			args: fmt.Sprintf("list --store-id %s --json", storeID),
			api: mock.API{
				ListSecretsFn: func(i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
					return secrets, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     encodeJSON(secrets),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameSecret+" "+testcase.args), &stdout)

			f := testcase.api.ListSecretsFn
			var apiInvoked bool
			testcase.api.ListSecretsFn = func(i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListSecrets invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
