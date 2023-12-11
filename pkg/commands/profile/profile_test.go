package profile_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

// Scenario is an extension of the base TestScenario.
// It includes manipulating stdin.
type Scenario struct {
	testutil.TestScenario

	ConfigFile config.File
	Stdin      []string
}

func TestProfileCreate(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name: "validate profile creation works",
				Args: args("profile create foo"),
				API: mock.API{
					GetTokenSelfFn: getToken,
					GetUserFn:      getUser,
				},
				WantOutputs: []string{
					"Fastly API token:",
					"Validating token",
					"Persisting configuration",
					"Profile 'foo' created",
				},
			},
			Stdin: []string{"some_token"},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate profile duplication",
				Args:      args("profile create foo"),
				WantError: "profile 'foo' already exists",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
				},
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			// TODO: abstract the logic for handling interactive stdin prompts.
			// This same if/else block is fundamentally duplicated across test files.
			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Input = stdin

				// Wait for user input and write it to the prompt
				inputc := make(chan string)
				go func() {
					for input := range inputc {
						fmt.Fprintln(prompt, input)
					}
				}()

				// We need a channel so we wait for `run()` to complete
				done := make(chan bool)

				// Call `app.Run()` and wait for response
				go func() {
					app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
						return opts, nil
					}
					err = app.Run(testcase.Args, nil)
					done <- true
				}()

				// User provides input
				//
				// NOTE: Must provide as much input as is expected to be waited on by `run()`.
				//       For example, if `run()` calls `input()` twice, then provide two messages.
				//       Otherwise the select statement will trigger the timeout error.
				for _, input := range testcase.Stdin {
					inputc <- input
				}

				select {
				case <-done:
					// Wait for app.Run() to finish
				case <-time.After(time.Second):
					t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
				}
			} else {
				stdin := ""
				if len(testcase.Stdin) > 0 {
					stdin = testcase.Stdin[0]
				}
				opts.Input = strings.NewReader(stdin)
				app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
					return opts, nil
				}
				err = app.Run(testcase.Args, nil)
			}

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestProfileDelete(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate profile deletion works",
				Args:       args("profile delete foo"),
				WantOutput: "Profile 'foo' deleted",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate incorrect profile",
				Args:      args("profile delete unknown"),
				WantError: "the specified profile does not exist",
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestProfileList(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name: "validate listing profiles works",
				Args: args("profile list"),
				WantOutputs: []string{
					"Default profile highlighted in red.",
					"foo\n\nDefault: true\nEmail: foo@example.com\nToken: 123",
					"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
				},
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate no profiles defined",
				Args:      args("profile list"),
				WantError: "no profiles available",
			},
		},
		// NOTE: The following test is subtly different to the previous one in that
		// our logic checks whether the config.Profiles map type is nil. If it is
		// then we error (see above test), otherwise if the map is set but there
		// are no profiles, then we notify the user no profiles exist.
		{
			TestScenario: testutil.TestScenario{
				Name: "validate no profiles available",
				Args: args("profile list"),
				WantOutputs: []string{
					"No profiles defined. To create a profile, run",
					"fastly profile create <name>",
				},
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate listing profiles displays warning if no default set",
				Args: args("profile list"),
				WantOutputs: []string{
					"At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>`.",
					"foo\n\nDefault: false\nEmail: foo@example.com\nToken: 123",
					"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
				},
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: false,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate listing profiles with --verbose and --json causes an error",
				Args:      args("profile list --verbose --json"),
				WantError: "invalid flag combination, --verbose and --json",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: false,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate listing profiles with --json displays data correctly",
				Args: args("profile list --json"),
				WantOutput: `{
  "bar": {
    "access_token": "",
    "access_token_created": 0,
    "access_token_ttl": 0,
    "default": false,
    "email": "bar@example.com",
    "refresh_token": "",
    "refresh_token_created": 0,
    "refresh_token_ttl": 0,
    "token": "456"
  },
  "foo": {
    "access_token": "",
    "access_token_created": 0,
    "access_token_ttl": 0,
    "default": false,
    "email": "foo@example.com",
    "refresh_token": "",
    "refresh_token_created": 0,
    "refresh_token_ttl": 0,
    "token": "123"
  }
}`,
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: false,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestProfileSwitch(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate switching to unknown profile returns an error",
				Args:      args("profile switch unknown"),
				WantError: "the profile 'unknown' does not exist",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate switching profiles works",
				Args:       args("profile switch bar"),
				WantOutput: "Profile switched to 'bar'",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestProfileToken(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate the active profile token is displayed by default",
				Args:       args("profile token"),
				WantOutput: "123",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate token is displayed for the specified profile",
				Args:       args("profile token bar"), // we choose a non-default profile
				WantOutput: "456",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate token is displayed for the specified profile using global --profile",
				Args:       args("profile token --profile bar"), // we choose a non-default profile
				WantOutput: "456",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate an unrecognised profile causes an error",
				Args:      args("profile token unknown"),
				WantError: "profile 'unknown' does not exist",
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestProfileUpdate(t *testing.T) {
	var (
		configPath string
		data       []byte
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// Read the test config.toml data
		path, err := filepath.Abs(filepath.Join("./", "testdata", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new test environment along with a test config.toml file.
		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(data), Dst: "config.toml"},
			},
		})
		configPath = filepath.Join(rootdir, "config.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name:      "validate updating unknown profile returns an error",
				Args:      args("profile update unknown"),
				WantError: "the profile 'unknown' does not exist",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Name: "validate updating profile works",
				Args: args("profile update bar"), // we choose a non-default profile
				API: mock.API{
					GetTokenSelfFn: getToken,
					GetUserFn:      getUser,
				},
				WantOutput: "Profile 'bar' updated",
			},
			ConfigFile: config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
					"bar": &config.Profile{
						Default: false,
						Email:   "bar@example.com",
						Token:   "456",
					},
				},
			},
			Stdin: []string{
				"",  // we skip SSO prompt
				"",  // we skip updating the token
				"y", // we set the profile to be the default
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.Config = testcase.ConfigFile

			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Input = stdin

				// Wait for user input and write it to the prompt
				inputc := make(chan string)
				go func() {
					for input := range inputc {
						fmt.Fprintln(prompt, input)
					}
				}()

				// We need a channel so we wait for `run()` to complete
				done := make(chan bool)

				// Call `app.Run()` and wait for response
				go func() {
					app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
						return opts, nil
					}
					err = app.Run(testcase.Args, nil)
					done <- true
				}()

				// User provides input
				//
				// NOTE: Must provide as much input as is expected to be waited on by `run()`.
				//       For example, if `run()` calls `input()` twice, then provide two messages.
				//       Otherwise the select statement will trigger the timeout error.
				for _, input := range testcase.Stdin {
					inputc <- input
				}

				select {
				case <-done:
					// Wait for app.Run() to finish
				case <-time.After(time.Second):
					t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
				}
			} else {
				stdin := ""
				if len(testcase.Stdin) > 0 {
					stdin = testcase.Stdin[0]
				}
				opts.Input = strings.NewReader(stdin)
				app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
					return opts, nil
				}
				err = app.Run(testcase.Args, nil)
			}

			t.Log(stdout.String())

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

func getUser(i *fastly.GetUserInput) (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		UserID:                 fastly.ToPointer(i.UserID),
		Login:                  fastly.ToPointer("foo@example.com"),
		Name:                   fastly.ToPointer("foo"),
		Role:                   fastly.ToPointer("user"),
		CustomerID:             fastly.ToPointer("abc"),
		EmailHash:              fastly.ToPointer("example-hash"),
		LimitServices:          fastly.ToPointer(true),
		Locked:                 fastly.ToPointer(true),
		RequireNewPassword:     fastly.ToPointer(true),
		TwoFactorAuthEnabled:   fastly.ToPointer(true),
		TwoFactorSetupRequired: fastly.ToPointer(true),
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}
