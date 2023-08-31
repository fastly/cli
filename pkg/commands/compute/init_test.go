package compute_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

func TestInit(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_INIT") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_INIT to run this test")
	}

	skRust := []config.StarterKit{
		{
			Name:   "Default",
			Path:   "https://github.com/fastly/compute-starter-kit-rust-default",
			Branch: "main",
		},
	}
	skJS := []config.StarterKit{
		{
			Name:   "Default",
			Path:   "https://github.com/fastly/compute-starter-kit-javascript-default",
			Branch: "main",
		},
	}

	scenarios := []struct {
		name             string
		args             []string
		configFile       config.File
		manifest         string
		wantFiles        []string
		unwantedFiles    []string
		wantError        string
		wantOutput       []string
		manifestIncludes string
		manifestPath     string
		stdin            string
	}{
		{
			name:      "broken endpoint",
			args:      args("compute init --from https://example.com/i-dont-exist"),
			wantError: "failed to get package: 404 Not Found",
		},
		{
			name: "name prompt",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			stdin: "foobar", // expect the first prompt to be for the package name.
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `name = "foobar"`,
		},
		{
			name: "description prompt empty",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `description = ""`, // expect this to be empty
		},
		{
			name: "with author",
			args: args("compute init --author test@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name: "with multiple authors",
			args: args("compute init --author test1@example.com --author test2@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},
		{
			name: "with --from set to starter kit repository",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to starter kit repository with .git extension and branch",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default.git --branch main"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to zip archive",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default/archive/refs/heads/main.zip"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to tar.gz archive",
			args: args("compute init --from https://github.com/Integralist/devnull/files/7339887/compute-starter-kit-rust-default-main.tar.gz"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with existing fastly.toml",
			args: args("compute init --auto-yes"), // --force will ignore a directory that isn't empty
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			manifest: `
			manifest_version = 2
			service_id = 1234
			name = "test"
			language = "rust"
			description = "test"
			authors = ["test@fastly.com"]`,
			wantOutput: []string{
				"Reading fastly.toml",
				"Initializing package",
			},
		},
		{
			name: "with existing fastly.toml",
			args: args("compute init --auto-yes"), // --force will ignore a directory that isn't empty
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			manifest: `
			manifest_version = 2
			service_id = 1234
			name = "test"
			language = "rust"
			description = "test"
			authors = ["test@fastly.com"]`,
			wantOutput: []string{
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "no args and no user profiles means no email set for author field",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantFiles: []string{
				"Cargo.toml",
				"fastly.toml",
				"src/main.rs",
			},
			unwantedFiles: []string{
				"SECURITY.md",
			},
			wantOutput: []string{
				"Author (email): Language:",
				"Fetching package template",
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "no args but email defaults to config.toml value in author field",
			args: args("compute init"),
			configFile: config.File{
				Profiles: config.Profiles{
					"user": &config.Profile{
						Email:   "test@example.com",
						Default: true,
					},
					"non_default": &config.Profile{
						Email: "no-default@example.com",
					},
				},
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
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
				"Fetching package template",
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "non empty directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantError: "project directory not empty",
			manifest: `
			manifest_version = 2
			name = "test"`,
		},
		{
			name: "with default name inferred from directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name: "with directory name inferred from --directory",
			args: args("compute init --directory ./foo"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			stdin:            "Y",
			manifest:         `manifest_version = 2`,
			manifestPath:     "foo",
			manifestIncludes: `name = "foo`,
		},
		{
			name: "with JavaScript language",
			args: args("compute init --language javascript"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					JavaScript: skJS,
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		// NOTE: This test verifies that we don't fetch a remote project.
		// Whether that be a starter kit or custom project template.
		// This is because "other" indicates an unsupported platform language.
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an init environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			manifestPath := filepath.Join(testcase.manifestPath, manifest.Filename)

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Write: []testutil.FileIO{
					{Src: testcase.manifest, Dst: manifestPath},
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ConfigFile = testcase.configFile

			// we need to define stdin as the init process prompts the user multiple
			// times, but we don't need to provide any values as all our prompts will
			// fallback to default values if the input is unrecognised.
			opts.Stdin = strings.NewReader(testcase.stdin)

			err = app.Run(opts)

			t.Log(stdout.String())

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
				content, err := os.ReadFile(filepath.Join(rootdir, manifestPath))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}
