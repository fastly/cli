package compute_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/testutil"
)

func TestInit(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_INIT") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_INIT to run this test")
	}

	for _, testcase := range []struct {
		name             string
		args             []string
		configFile       config.File
		manifest         string
		wantFiles        []string
		unwantedFiles    []string
		wantError        string
		wantOutput       []string
		manifestIncludes string
	}{
		{
			name:      "unknown repository",
			args:      args("compute init --from https://example.com/template"),
			wantError: "error fetching package template:",
		},
		{
			name: "with name",
			args: args("compute init --name test"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `name = "test"`,
		},
		{
			name: "with description",
			args: args("compute init --description test"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `description = "test"`,
		},
		{
			name: "with author",
			args: args("compute init --author test@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name: "with multiple authors",
			args: args("compute init --author test1@example.com --author test2@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},
		{
			name: "with from repository and branch",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default.git --branch main"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name: "with existing package manifest",
			args: args("compute init --force"), // --force will ignore a directory that isn't empty
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifest: strings.Join([]string{
				"manifest_version = \"1\"",
				"service_id = \"1234\"",
				"name = \"test\"",
				"language = \"rust\"",
				"description = \"test\"",
				"authors = [\"test@fastly.com\"]",
			}, "\n"),
			wantOutput: []string{
				"Updating package manifest...",
				"Initializing package...",
			},
		},
		{
			name: "default",
			args: args("compute init"),
			configFile: config.File{
				User: config.User{
					Email: "test@example.com",
				},
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifestIncludes: `authors = ["test@example.com"]`,
			wantFiles: []string{
				"Cargo.toml",
				"fastly.toml",
				"src/main.rs",
			},
			unwantedFiles: []string{
				"SECURITY.md",
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name: "non empty directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantError: "project directory not empty",
			manifest:  `name = "test"`, // causes a file to be created as part of test setup
		},
		{
			name: "with default name inferred from directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name: "with AssemblyScript language",
			args: args("compute init --language assemblyscript"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					AssemblyScript: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-assemblyscript-default",
							Tag:  "v0.2.0",
						},
					},
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name:             "with pre-compiled Wasm binary",
			args:             args("compute init --language other"),
			manifestIncludes: `language = "other"`,
			wantOutput: []string{
				"Initialized package",
				"To package a pre-compiled Wasm binary for deployment",
				"SUCCESS: Initialized package",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an init environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Write: []testutil.FileIO{
					{Src: testcase.manifest, Dst: manifest.Filename},
				},
			})
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the init environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetFile(testcase.configFile)

			// we need to define stdin as the init process prompts the user multiple
			// times, but we don't need to provide any values as all our prompts will
			// fallback to default values if the input is unrecognised.
			ara.SetStdin(strings.NewReader(""))

			err = app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, file := range testcase.wantFiles {
				if _, err := os.Stat(filepath.Join(rootdir, file)); err != nil {
					t.Errorf("wanted file %s not found", file)
				}
			}
			for _, file := range testcase.unwantedFiles {
				if _, err := os.Stat(filepath.Join(rootdir, file)); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("unwanted file %s found", file)
				}
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, manifest.Filename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}
