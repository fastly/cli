package config_test

import (
	"bytes"
	_ "embed"
	"strings"

	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
	toml "github.com/pelletier/go-toml"
)

type mockHTTPClient struct{}

func (c mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, context.DeadlineExceeded
}

// TestConfigLoad validates that when a context.DeadlineExceeded error is
// returned from a http.Client.Do request, that an appropriate remediation
// error is returned to the user.
func TestConfigLoad(t *testing.T) {
	var (
		c mockHTTPClient
		d time.Duration
		f *config.File
	)
	if err := f.Load("foo", c, d, "/path/to/config.toml"); err != nil {
		if !errors.As(err, &fsterr.RemediationError{}) {
			t.Errorf("expected RemediationError got: %T", err)
		}
		if err.(fsterr.RemediationError).Remediation != fsterr.NetworkRemediation {
			t.Errorf("expected NetworkRemediation got: %s", err.(fsterr.RemediationError).Remediation)
		}
	} else {
		t.Error("expected an error, got nil")
	}
}

//go:embed testdata/static/config.toml
var staticConfig []byte

//go:embed testdata/static/config-invalid.toml
var staticConfigInvalid []byte

type testReadScenario struct {
	lastCheckedVersionSet bool
	name                  string
	remediation           bool
	staticConfig          []byte
	userConfigFilename    string
	userResponseToPrompt  string
	wantError             string
}

// TestConfigRead validates all of the logic flow within config.File.Read()
func TestConfigRead(t *testing.T) {
	scenarios := []testReadScenario{
		{
			name:                  "static config should be used when there is no local user config",
			lastCheckedVersionSet: true,
			staticConfig:          staticConfig,
		},
		{
			name:         "static config should return an error when invalid",
			staticConfig: staticConfigInvalid,
			wantError:    config.ErrInvalidConfig.Error(),
		},
		{
			name:                 "when user config is invalid (and the user accepts static config) but static config is also invalid, it should return an error",
			staticConfig:         staticConfigInvalid,
			userConfigFilename:   "config-invalid.toml",
			userResponseToPrompt: "yes",
			wantError:            config.ErrInvalidConfig.Error(),
		},
		{
			name:                 "when user config is invalid (and the user rejects static config), it should return a specific remediation error",
			remediation:          true,
			staticConfig:         staticConfig,
			userConfigFilename:   "config-invalid.toml",
			userResponseToPrompt: "no",
			wantError:            config.RemediationManualFix,
		},
		{
			name:                  "when user config is in the legacy format, it should return a specific error",
			lastCheckedVersionSet: true,
			staticConfig:          staticConfig,
			userConfigFilename:    "config-legacy.toml",
			userResponseToPrompt:  "no",
			wantError:             config.ErrLegacyConfig.Error(),
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an temp environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our environment in a temp dir.
			// Defer a call to clean it up.
			rootdir, fpath := makeTempEnvironment(t, testcase.userConfigFilename)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the temp environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			if testcase.userConfigFilename == "" {
				if fi, err := os.Stat(fpath); err == nil {
					t.Fatalf("expected the user config to NOT exist at this point: %+v", fi)
				}
			} else {
				if _, err := os.Stat(fpath); err != nil {
					t.Fatalf("expected the user config to exist at this point: %v", err)
				}
			}

			var f config.File
			f.Static = testcase.staticConfig

			var out bytes.Buffer
			in := strings.NewReader(testcase.userResponseToPrompt)

			err = f.Read(fpath, in, &out)

			if testcase.remediation {
				e, ok := err.(fsterr.RemediationError)
				if !ok {
					t.Fatalf("unexpected error asserting returned error (%T) to a RemediationError type", err)
				}
				if testcase.wantError != e.Remediation {
					t.Fatalf("want %v, have %v", testcase.wantError, e.Remediation)
				}
			} else {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			}

			if testcase.wantError == "" {
				// If we're not expecting an error, then we're expecting the user
				// configuration file to exist...

				if _, err := os.Stat(fpath); err == nil {
					bs, err := os.ReadFile(fpath)
					if err != nil {
						t.Fatalf("unexpected err: %v", err)
					}

					err = toml.Unmarshal(bs, &f)
					if err != nil {
						t.Fatalf("unexpected err: %v", err)
					}

					if f.CLI.LastChecked == "" || f.CLI.Version == "" {
						t.Fatalf("expected LastChecked/Version to be set: %+v", f)
					}
				}
			}
		})
	}
}

// TestUseStatic validates all of the logic flow within config.File.UseStatic()
func TestUseStatic(t *testing.T) {
	// We're going to chdir to an temp environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create our environment in a temp dir.
	// Defer a call to clean it up.
	rootdir, fpath := makeTempEnvironment(t, "config-legacy.toml")
	defer os.RemoveAll(rootdir)

	// Before running the test, chdir into the temp environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably assert file structure.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	f := config.File{}

	// Validate that invalid static configuration returns a specific error.
	err = f.UseStatic(staticConfigInvalid, fpath)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	} else {
		testutil.AssertErrorContains(t, err, config.ErrInvalidConfig.Error())
	}

	// Validate that legacy configuration can be migrated to the static one
	// embedded in the CLI binary.
	var out bytes.Buffer
	f.Read(fpath, strings.NewReader(""), &out)
	f.UseStatic(staticConfig, fpath)
	if f.CLI.LastChecked == "" || f.CLI.Version == "" {
		t.Fatalf("expected LastChecked/Version to be set: %+v", f)
	}
	if f.User.Token != "foobar" {
		t.Fatalf("wanted token: %s, got: %s", "foobar", f.User.Token)
	}
	if f.User.Email != "testing@fastly.com" {
		t.Fatalf("wanted email: %s, got: %s", "testing@fastly.com", f.User.Email)
	}
}

// TestConfigWrite validates all of the logic flow within config.File.Write()
//
// Specifically we're interested in whether the f.Static field is written to
// disk (we don't we want it to be) and whether its value is reset (we do want).
func TestConfigWrite(t *testing.T) {
	// We're going to chdir to an temp environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create our environment in a temp dir.
	// Defer a call to clean it up.
	rootdir, fpath := makeTempEnvironment(t, "config.toml")
	defer os.RemoveAll(rootdir)

	// Before running the test, chdir into the temp environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably assert file structure.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	// Validate the f.Static field is reset and restored.
	var f config.File
	f.Static = staticConfig
	f.Write(fpath)
	if len(f.Static) < 1 {
		t.Fatal("expected f.Static to be set")
	}

	// Validate the f.Static isn't written back to disk by reading from disk and
	// unmarshalling the data to check f.Static is now zero length.
	//
	// NOTE: We have to manually reset f.Static because toml.Unmarshal won't
	// reset any fields already with values (f.Static would still have a value
	// set as the field would have been reset at the end of the f.Write, which we
	// validated in the above test assertion).
	//
	// So although we manually reset f.Static, if the local toml file was
	// incorrectly marshalled with a static field, we'd expect the contents to be
	// set back into f.Static when we unmarshal the local toml file. It's that
	// behaviour of f.Write that we want to validate doesn't happen.
	f.Static = []byte("")
	bs, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = toml.Unmarshal(bs, &f)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(f.Static) > 0 {
		t.Fatalf("expected f.Static to be not set: %+v", f.Static)
	}
}

// makeTempEnvironment creates a temporary directory, and within that directory
// it creates a config.toml based on the src content from ./testdata/. If the
// calling test function passes an empty string, then not local user config is
// created (allowing for a test function to validate the static embedded one).
func makeTempEnvironment(t *testing.T, userConfigSrcFilename string) (rootdir, userConfig string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-temp-*")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(rootdir, 0700); err != nil {
		t.Fatal(err)
	}

	userConfig = filepath.Join(rootdir, "config.toml")
	configSrc, err := filepath.Abs(filepath.Join("testdata", userConfigSrcFilename))
	if err != nil {
		t.Fatal(err)
	}

	if userConfigSrcFilename != "" {
		b, err := os.ReadFile(configSrc)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(userConfig, b, 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir, userConfig
}
