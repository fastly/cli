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
	name                 string
	remediation          bool
	staticConfig         []byte
	userConfigFilename   string
	userResponseToPrompt string
	wantError            string
}

// TestConfigRead validates all logic flows within config.File.Read()
func TestConfigRead(t *testing.T) {
	scenarios := []testReadScenario{
		{
			name:         "static config should be used when there is no local user config",
			staticConfig: staticConfig,
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
			name:                 "when user config is in the legacy format, it should return a specific error",
			staticConfig:         staticConfig,
			userConfigFilename:   "config-legacy.toml",
			userResponseToPrompt: "no",
			wantError:            config.ErrLegacyConfig.Error(),
		},
		{
			name:               "when user config is valid, it should return no error",
			staticConfig:       staticConfig,
			userConfigFilename: "config.toml",
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

			// Create test environment
			opts := testutil.EnvOpts{T: t}
			if testcase.userConfigFilename != "" {
				b, err := os.ReadFile(filepath.Join("testdata", testcase.userConfigFilename))
				if err != nil {
					t.Fatal(err)
				}
				opts.Write = []testutil.FileIO{
					{Src: string(b), Dst: "config.toml"},
				}
			}
			rootdir := testutil.NewEnv(opts)
			configPath := filepath.Join(rootdir, "config.toml")
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the temp environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			if testcase.userConfigFilename == "" {
				if fi, err := os.Stat(configPath); err == nil {
					t.Fatalf("expected the user config to NOT exist at this point: %+v", fi)
				}
			} else {
				if _, err := os.Stat(configPath); err != nil {
					t.Fatalf("expected the user config to exist at this point: %v", err)
				}
			}

			var f config.File
			f.Static = testcase.staticConfig

			var out bytes.Buffer
			in := strings.NewReader(testcase.userResponseToPrompt)

			err = f.Read(configPath, in, &out)

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

				if _, err := os.Stat(configPath); err == nil {
					bs, err := os.ReadFile(configPath)
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

type testUseStaticScenario struct {
	name               string
	userConfigFilename string
}

// TestUseStatic validates all logic flows within config.File.UseStatic()
func TestUseStatic(t *testing.T) {
	scenarios := []testUseStaticScenario{
		{
			name:               "legacy config should be safe to migrate to static",
			userConfigFilename: "config-legacy.toml",
		},
		// The following scenario is specifically validating that if a recent
		// config (which still had a 'Legacy' section) was identified as needing to
		// be migrated to the static config embedded in the CLI (because the
		// user config's CLI.Version field didn't align with the currently running
		// CLI binary version), then it should still be transitioned safely without
		// losing the user's token/email within the 'User' section.
		{
			name:               "config with 'legacy' section should be safe to migrate to static",
			userConfigFilename: "config-with-old-legacy-section.toml",
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

			// Create test environment
			b, err := os.ReadFile(filepath.Join("testdata", testcase.userConfigFilename))
			if err != nil {
				t.Fatal(err)
			}
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Write: []testutil.FileIO{
					{Src: string(b), Dst: "config.toml"},
				},
			})
			configPath := filepath.Join(rootdir, "config.toml")
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the temp environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			// Validate that invalid static configuration returns a specific error.
			f := config.File{}
			err = f.UseStatic(staticConfigInvalid, configPath)
			if err == nil {
				t.Fatal("expected an error, but got nil")
			} else {
				testutil.AssertErrorContains(t, err, config.ErrInvalidConfig.Error())
			}

			// Validate that legacy configuration can be migrated to the static one
			// embedded in the CLI binary.
			f = config.File{}
			var out bytes.Buffer
			f.Read(configPath, strings.NewReader(""), &out)
			f.UseStatic(staticConfig, configPath)
			if f.CLI.LastChecked == "" || f.CLI.Version == "" {
				t.Fatalf("expected LastChecked/Version to be set: %+v", f)
			}
			if f.User.Token != "foobar" {
				t.Fatalf("wanted token: %s, got: %s", "foobar", f.User.Token)
			}
			if f.User.Email != "testing@fastly.com" {
				t.Fatalf("wanted email: %s, got: %s", "testing@fastly.com", f.User.Email)
			}
		})
	}
}

// TestConfigWrite validates all logic flows within config.File.Write()
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

	// Create test environment
	b, err := os.ReadFile(filepath.Join("testdata", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Write: []testutil.FileIO{
			{Src: string(b), Dst: "config.toml"},
		},
	})
	configPath := filepath.Join(rootdir, "config.toml")
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
	f.Write(configPath)
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
	bs, err := os.ReadFile(configPath)
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

type testValidConfigScenario struct {
	testutil.TestScenario

	ok            bool
	staticConfig  []byte
	userConfig    string
	verboseOutput bool
}

// TestValidConfig validates all logic flows within config.File.ValidConfig()
func TestValidConfig(t *testing.T) {
	s1 := testValidConfigScenario{}
	s1.Name = "invalid config"
	s1.staticConfig = staticConfig
	s1.userConfig = "config-incompatible-config-version.toml"

	s2 := testValidConfigScenario{}
	s2.Name = "invalid config with verbose output"
	s2.staticConfig = staticConfig
	s2.userConfig = "config-incompatible-config-version.toml"
	s2.verboseOutput = true

	s3 := testValidConfigScenario{}
	s3.Name = "valid config"
	s3.ok = true
	s3.staticConfig = staticConfig
	s3.userConfig = "config.toml"

	scenarios := []testValidConfigScenario{s1, s2, s3}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			// We're going to chdir to an temp environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Copy: []testutil.FileIO{
					{
						Src: filepath.Join("testdata", testcase.userConfig),
						Dst: filepath.Join("config.toml"),
					},
				},
			})
			configPath := filepath.Join(rootdir, "config.toml")
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the temp environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var f config.File
			f.Static = testcase.staticConfig

			var stdout bytes.Buffer
			in := strings.NewReader("") // these tests won't trigger a user prompt

			err = f.Read(configPath, in, &stdout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ok := f.ValidConfig(testcase.verboseOutput, &stdout)
			if ok != testcase.ok {
				t.Fatalf("want %t, have: %t", testcase.ok, ok)
			}

			output := strings.ReplaceAll(stdout.String(), "\n", " ")
			if !testcase.ok && testcase.verboseOutput {
				testutil.AssertStringContains(t, output, "incompatible with the current CLI version")
			}
		})
	}
}
