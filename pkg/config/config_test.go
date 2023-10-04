package config_test

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
)

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
			name:                 "static config should be used when there is no local user config",
			userResponseToPrompt: "yes", // prompts asks user to confirm they want a static fallback
			staticConfig:         staticConfig,
		},
		{
			name:                 "static config should return an error when invalid",
			userResponseToPrompt: "yes", // prompts asks user to confirm they want a static fallback
			staticConfig:         staticConfigInvalid,
			wantError:            config.ErrInvalidConfig.Error(),
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
			name:                 "when user config is in the legacy format, it should use static config",
			staticConfig:         staticConfig,
			userConfigFilename:   "config-legacy.toml",
			userResponseToPrompt: "no",
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
			backupStatic := config.Static
			defer func() {
				config.Static = backupStatic
			}()
			config.Static = testcase.staticConfig
			opts := testutil.EnvOpts{T: t}
			if testcase.userConfigFilename != "" {
				b, err := os.ReadFile(filepath.Join("testdata", testcase.userConfigFilename))
				if err != nil {
					t.Fatal(err)
				}
				opts.Write = []testutil.FileIO{
					{Src: string(b), Dst: "user-config.toml"},
				}
			}
			rootdir := testutil.NewEnv(opts)
			configPath := filepath.Join(rootdir, "user-config.toml")
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

			var out bytes.Buffer
			in := strings.NewReader(testcase.userResponseToPrompt)

			mockLog := fsterr.MockLog{}

			var f config.File
			err = f.Read(configPath, in, &out, mockLog, false)

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

					if f.CLI.Version == "" {
						t.Fatalf("expected CLI.Version to be set: %+v", f)
					}
				}
			}
		})
	}
}

// TestUseStatic validates legacy user data is migrated successfully.
func TestUseStatic(t *testing.T) {
	// We're going to chdir to an temp environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	b, err := os.ReadFile(filepath.Join("testdata", "config-legacy.toml"))
	if err != nil {
		t.Fatal(err)
	}
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Write: []testutil.FileIO{
			{Src: string(b), Dst: "user-config.toml"},
		},
	})
	legacyUserConfigPath := filepath.Join(rootdir, "user-config.toml")
	defer os.RemoveAll(rootdir)

	// Before running the test, chdir into the temp environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably assert file structure.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	var out bytes.Buffer

	// Validate that legacy configuration can be migrated to the static one
	// embedded in the CLI binary.
	f := config.File{}
	err = f.Read(legacyUserConfigPath, strings.NewReader(""), &out, fsterr.MockLog{}, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if f.CLI.Version == "" {
		t.Fatalf("expected CLI.Version to be set: %+v", f)
	}
	if f.Profiles["user"].Token != "foobar" {
		t.Fatalf("wanted token: %s, got: %s", "foobar", f.LegacyUser.Token)
	}
	if f.Profiles["user"].Email != "testing@fastly.com" {
		t.Fatalf("wanted email: %s, got: %s", "testing@fastly.com", f.LegacyUser.Email)
	}
	if !f.Profiles["user"].Default {
		t.Fatal("expected the migrated user to become the default")
	}

	// We validate both the in-memory data structure (above) AND the file on disk (below).
	data, err := os.ReadFile(legacyUserConfigPath)
	if err != nil {
		t.Error(err)
	}
	if strings.Contains(string(data), "[user]") {
		t.Error("expected legacy [user] section to be removed")
	}
	if !strings.Contains(string(data), "[profile.user]\ndefault = true\nemail = \"testing@fastly.com\"\ntoken = \"foobar\"") {
		t.Error("expected legacy [user] section to be migrated to [profile.user]")
	}

	// Validate that invalid static configuration returns a specific error.
	//
	// NOTE: By providing a legacy config, we'll cause the static config embedded
	// into the CLI to be used, and we'll migrate the legacy data to the new
	// format, but by specifying the static config as being invalid we expect the
	// CLI to return the error.
	backupStatic := config.Static
	defer func() {
		config.Static = backupStatic
	}()
	config.Static = staticConfigInvalid
	f = config.File{}
	err = f.Read(legacyUserConfigPath, strings.NewReader(""), &out, fsterr.MockLog{}, false)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	} else {
		testutil.AssertErrorContains(t, err, config.ErrInvalidConfig.Error())
	}
}

type testInvalidConfigScenario struct {
	testutil.TestScenario

	invalid      bool
	staticConfig []byte
	userConfig   string
}

// TestInvalidConfig validates all logic flows within config.File.ValidConfig()
//
// NOTE: Even with invalid config we expect the static config embedded with the
// CLI to be utilised.
func TestInvalidConfig(t *testing.T) {
	s1 := testInvalidConfigScenario{}
	s1.Name = "invalid config version, invalid cli version"
	s1.invalid = true
	s1.staticConfig = staticConfig
	s1.userConfig = "config-incompatible-config-version.toml"

	s2 := testInvalidConfigScenario{}
	s2.Name = "valid config version, invalid cli version"
	s2.invalid = false
	s2.staticConfig = staticConfig
	s2.userConfig = "config.toml"

	scenarios := []testInvalidConfigScenario{s1, s2}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
			var stdout bytes.Buffer
			config.Static = testcase.staticConfig

			in := strings.NewReader("") // these tests won't trigger a user prompt
			err = f.Read(configPath, in, &stdout, nil, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := strings.ReplaceAll(stdout.String(), "\n", " ")
			if testcase.invalid {
				testutil.AssertStringContains(t, output, "incompatible with the current CLI version")
			}
		})
	}
}

func TestNeedsUpdating(t *testing.T) {
	t.Parallel()

	config.CurrentConfigVersion = 2

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			"legacy config should be updated",
			"config-legacy.toml",
			true,
		},
		{
			"outdated config_version config should be updated",
			"config.toml",
			true,
		},
		{
			"mismatching CLI version config should be updated",
			"config-outdated-cli-version.toml",
			true,
		},
		{
			"current config should not be updated",
			"config-current.toml",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var f config.File
			if err = toml.Unmarshal(data, &f); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var stdout bytes.Buffer
			mockLog := fsterr.MockLog{}
			result := f.NeedsUpdating(data, &stdout, mockLog, true)
			if result != tt.want {
				t.Fatalf("expected %v got %v", tt.want, result)
			}
		})
	}
}
