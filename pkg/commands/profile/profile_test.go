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

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

// Scenario is an extension of the base TestScenario.
// It includes manipulating stdin.
type Scenario struct {
	testutil.TestScenario

	ConfigFile config.File
	Stdin      []string
}

func TestCreate(t *testing.T) {
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
		defer os.Chdir(wd)
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
					"Validating token...",
					"Persisting configuration...",
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

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			// TODO: abstract the logic for handling interactive stdin prompts.
			// This same if/else block is fundamentally duplicated across test files.
			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Stdin = stdin

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
					err = app.Run(opts)
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
				opts.Stdin = strings.NewReader(stdin)
				err = app.Run(opts)
			}

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestDelete(t *testing.T) {
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
		defer os.Chdir(wd)
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

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestList(t *testing.T) {
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
		defer os.Chdir(wd)
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
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestSwitch(t *testing.T) {
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
		defer os.Chdir(wd)
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

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestToken(t *testing.T) {
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
		defer os.Chdir(wd)
	}

	args := testutil.Args
	scenarios := []Scenario{
		{
			TestScenario: testutil.TestScenario{
				Name:       "validate default user token is displayed",
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
				Name:       "validate specified user token is displayed",
				Args:       args("profile token --user bar"), // we choose a non-default profile
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
				Name:      "validate unknown user causes an error",
				Args:      args("profile token --user unknown"),
				WantError: "profile 'unknown' does not exist",
			},
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestUpdate(t *testing.T) {
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
		defer os.Chdir(wd)
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
				"",  // we skip updating the token
				"y", // we set the profile to be the default
			},
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				err    error
				stdout bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// The read of the config file only really happens in the main()
			// function, so for the sake of the test environment we need to construct
			// an in-memory representation of the config file we want to be using.
			opts.ConfigFile = testcase.ConfigFile

			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Stdin = stdin

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
					err = app.Run(opts)
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
				opts.Stdin = strings.NewReader(stdin)
				err = app.Run(opts)
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
		ID:         "123",
		Name:       "Foo",
		UserID:     "456",
		Services:   []string{"a", "b"},
		Scope:      fastly.TokenScope(fmt.Sprintf("%s %s", fastly.PurgeAllScope, fastly.GlobalReadScope)),
		IP:         "127.0.0.1",
		CreatedAt:  &t,
		ExpiresAt:  &t,
		LastUsedAt: &t,
	}, nil
}

func getUser(i *fastly.GetUserInput) (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		ID:                     i.ID,
		Login:                  "foo@example.com",
		Name:                   "foo",
		Role:                   "user",
		CustomerID:             "abc",
		EmailHash:              "example-hash",
		LimitServices:          true,
		Locked:                 true,
		RequireNewPassword:     true,
		TwoFactorAuthEnabled:   true,
		TwoFactorSetupRequired: true,
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}
