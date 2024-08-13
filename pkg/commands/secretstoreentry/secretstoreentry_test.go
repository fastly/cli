package secretstoreentry_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"
	"golang.org/x/crypto/nacl/box"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/secretstoreentry"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
	if err := os.WriteFile(secretFile, []byte(secretValue), 0o600); err != nil {
		t.Fatal(err)
	}
	doesNotExistFile := path.Join(tmpDir, "DOES-NOT-EXIST")

	ckPub, ckPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	skPub, skPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	ck := &fastly.ClientKey{
		PublicKey: ckPub[:],
		Signature: ed25519.Sign(skPriv, ckPub[:]),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockCreateClientKey := func() (*fastly.ClientKey, error) { return ck, nil }
	mockGetSigningKey := func() (ed25519.PublicKey, error) { return skPub, nil }

	decrypt := func(ciphertext []byte) (string, error) {
		plaintext, ok := box.OpenAnonymous(nil, ciphertext, ckPub, ckPriv)
		if !ok {
			return "", errors.New("failed to decrypt")
		}
		return string(plaintext), nil
	}

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
		{
			args:      fmt.Sprintf("create --store-id %s --name %s --stdin --recreate --recreate-allow", storeID, secretName),
			wantError: "invalid flag combination, --recreate and --recreate-allow",
		},
		// Read from STDIN.
		{
			args:  fmt.Sprintf("create --store-id %s --name %s --stdin", storeID, secretName),
			stdin: secretValue,
			api: mock.API{
				CreateClientKeyFn: mockCreateClientKey,
				GetSigningKeyFn:   mockGetSigningKey,
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if got, err := decrypt(i.Secret); err != nil {
						return nil, err
					} else if got != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", got)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Created secret '%s' in Secret Store '%s' (digest: %s)", secretName, storeID, hex.EncodeToString([]byte(secretDigest))),
		},
		// Read from file.
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s", storeID, secretName, secretFile),
			api: mock.API{
				CreateClientKeyFn: mockCreateClientKey,
				GetSigningKeyFn:   mockGetSigningKey,
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if got, err := decrypt(i.Secret); err != nil {
						return nil, err
					} else if got != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", got)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Created secret '%s' in Secret Store '%s' (digest: %s)", secretName, storeID, hex.EncodeToString([]byte(secretDigest))),
		},
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s --json", storeID, secretName, secretFile),
			api: mock.API{
				CreateClientKeyFn: mockCreateClientKey,
				GetSigningKeyFn:   mockGetSigningKey,
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if got, err := decrypt(i.Secret); err != nil {
						return nil, err
					} else if got != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", got)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.Secret{
				Name:   secretName,
				Digest: []byte(secretDigest),
			}),
		},
		// CreateOrRecreate
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s --json --recreate-allow", storeID, secretName, secretFile),
			api: mock.API{
				CreateClientKeyFn: mockCreateClientKey,
				GetSigningKeyFn:   mockGetSigningKey,
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if got, want := i.Method, http.MethodPut; got != want {
						return nil, fmt.Errorf("got method %q, want %q", got, want)
					}
					if got, err := decrypt(i.Secret); err != nil {
						return nil, err
					} else if got != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", got)
					}
					return &fastly.Secret{
						Name:   i.Name,
						Digest: []byte(secretDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.Secret{
				Name:   secretName,
				Digest: []byte(secretDigest),
			}),
		},
		// Recreate
		{
			args: fmt.Sprintf("create --store-id %s --name %s --file %s --json --recreate", storeID, secretName, secretFile),
			api: mock.API{
				CreateClientKeyFn: mockCreateClientKey,
				GetSigningKeyFn:   mockGetSigningKey,
				CreateSecretFn: func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
					if got, want := i.Method, http.MethodPatch; got != want {
						return nil, fmt.Errorf("got method %q, want %q", got, want)
					}
					if got, err := decrypt(i.Secret); err != nil {
						return nil, err
					} else if got != secretValue {
						return nil, fmt.Errorf("invalid secret: %s", got)
					}
					return &fastly.Secret{
						Name:      i.Name,
						Digest:    []byte(secretDigest),
						Recreated: true,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.Secret{
				Name:      secretName,
				Digest:    []byte(secretDigest),
				Recreated: true,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstoreentry.RootNameSecret + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)
			if testcase.stdin != "" {
				var stdin bytes.Buffer
				stdin.WriteString(testcase.stdin)
				opts.Input = &stdin
			}

			f := testcase.api.CreateSecretFn
			var apiInvoked bool
			testcase.api.CreateSecretFn = func(i *fastly.CreateSecretInput) (*fastly.Secret, error) {
				apiInvoked = true
				return f(i)
			}

			// Tests generate their own signing keys, which won't match
			// the hardcoded value.  Disable the check against the
			// hardcoded value.
			t.Setenv("FASTLY_USE_API_SIGNING_KEY", "1")

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
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
					if i.StoreID != storeID || i.Name != secretName {
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
					if i.StoreID != storeID || i.Name != secretName {
						return errors.New("not found")
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Deleted secret '%s' from Secret Store '%s'", secretName, storeID),
		},
		{
			args: fmt.Sprintf("delete --store-id %s --name %s --json", storeID, secretName),
			api: mock.API{
				DeleteSecretFn: func(i *fastly.DeleteSecretInput) error {
					if i.StoreID != storeID || i.Name != secretName {
						return errors.New("not found")
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.JSON(`{"name": %q, "store_id": %q,  "deleted": true}`, secretName, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstoreentry.RootNameSecret + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.DeleteSecretFn
			var apiInvoked bool
			testcase.api.DeleteSecretFn = func(i *fastly.DeleteSecretInput) error {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDescribeSecretCommand(t *testing.T) {
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
					if i.StoreID != storeID || i.Name != storeName {
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
					if i.StoreID != storeID || i.Name != storeName {
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
					if i.StoreID != storeID || i.Name != storeName {
						return nil, errors.New("invalid request")
					}
					return &fastly.Secret{
						Name:   storeName,
						Digest: []byte(storeDigest),
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.Secret{
				Name:   storeName,
				Digest: []byte(storeDigest),
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstoreentry.RootNameSecret + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.GetSecretFn
			var apiInvoked bool
			testcase.api.GetSecretFn = func(i *fastly.GetSecretInput) (*fastly.Secret, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API GetSecret invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
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
			wantOutput:     fstfmt.EncodeJSON(secrets),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstoreentry.RootNameSecret + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.ListSecretsFn
			var apiInvoked bool
			testcase.api.ListSecretsFn = func(i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListSecrets invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
