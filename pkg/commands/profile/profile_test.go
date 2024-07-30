package profile_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const ()

// Create temp environment to run test code within.
func createTempEnvironment(t *testing.T) (string, string) {
	var data []byte

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

	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}

	return filepath.Join(rootdir, "config.toml"), rootdir
}

func TestProfileCreate(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name: "validate profile creation works",
			Arg:  "foo",
			API: mock.API{
				GetTokenSelfFn: getToken,
				GetUserFn:      getUser,
			},
			ConfigPath: configPath,
			Stdin:      []string{"some_token"},
			WantOutputs: []string{
				"Fastly API token:",
				"Validating token",
				"Persisting configuration",
				"Profile 'foo' created",
			},
		},
		{
			Name:       "validate profile duplication",
			Arg:        "foo",
			WantError:  "profile 'foo' already exists",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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

	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestProfileDelete(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name:       "validate profile deletion works",
			Arg:        "foo",
			WantOutput: "Profile 'foo' deleted",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate incorrect profile",
			Arg:        "unknown",
			ConfigPath: configPath,
			WantError:  "the specified profile does not exist",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestProfileList(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name: "validate listing profiles works",
			WantOutputs: []string{
				"Default profile highlighted in red.",
				"foo\n\nDefault: true\nEmail: foo@example.com\nToken: 123",
				"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
			},
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate no profiles defined",
			ConfigPath: configPath,
			ConfigFile: &config.File{},
			WantError:  "no profiles available",
		},
		// NOTE: The following test is subtly different to the previous one in that
		// our logic checks whether the config.Profiles map type is nil. If it is
		// then we error (see above test), otherwise if the map is set but there
		// are no profiles, then we notify the user no profiles exist.
		{
			Name: "validate no profiles available",
			WantOutputs: []string{
				"No profiles defined. To create a profile, run",
				"fastly profile create <name>",
			},
			ConfigPath: configPath,
			ConfigFile: &config.File{
				Profiles: config.Profiles{},
			},
		},
		{
			Name: "validate listing profiles displays warning if no default set",
			WantOutputs: []string{
				"At least one account profile should be set as the 'default'.",
				"foo\n\nDefault: false\nEmail: foo@example.com\nToken: 123",
				"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
			},
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate listing profiles with --verbose and --json causes an error",
			Arg:        "--verbose --json",
			WantError:  "invalid flag combination, --verbose and --json",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name: "validate listing profiles with --json displays data correctly",
			Arg:  "--json",
			WantOutput: `{
  "bar": {
    "access_token": "",
    "access_token_created": 0,
    "access_token_ttl": 0,
    "customer_id": "",
    "customer_name": "",
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
    "customer_id": "",
    "customer_name": "",
    "default": false,
    "email": "foo@example.com",
    "refresh_token": "",
    "refresh_token_created": 0,
    "refresh_token_ttl": 0,
    "token": "123"
  }
}`,
			ConfigPath: configPath,
			ConfigFile: &config.File{
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

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestProfileSwitch(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name:       "validate switching to unknown profile returns an error",
			Arg:        "unknown",
			ConfigPath: configPath,
			WantError:  "the profile 'unknown' does not exist",
		},
		{
			Name:       "validate switching profiles works",
			Arg:        "bar",
			WantOutput: "Profile switched to 'bar'",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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

	testutil.RunScenarios(t, []string{root.CommandName, "switch"}, scenarios)
}

func TestProfileToken(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name:       "validate the active profile token is displayed by default",
			WantOutput: "123",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate token is displayed for the specified profile",
			Arg:        "bar", // we choose a non-default profile
			WantOutput: "456",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate token is displayed for the specified profile using global --profile",
			Arg:        "--profile bar", // we choose a non-default profile
			WantOutput: "456",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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
			Name:       "validate an unrecognised profile causes an error",
			Arg:        "unknown",
			WantError:  "profile 'unknown' does not exist",
			ConfigPath: configPath,
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "token"}, scenarios)
}

func TestProfileUpdate(t *testing.T) {
	var configPath string

	{
		var rootdir string

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		configPath, rootdir = createTempEnvironment(t)
		defer os.RemoveAll(rootdir)

		defer func() {
			_ = os.Chdir(wd)
		}()
	}

	scenarios := []testutil.TestScenario{
		{
			Name:       "validate updating unknown profile returns an error",
			Arg:        "unknown",
			ConfigPath: configPath,
			WantError:  "the profile 'unknown' does not exist",
		},
		{
			Name: "validate updating profile works",
			Arg:  "bar", // we choose a non-default profile
			API: mock.API{
				GetTokenSelfFn: getToken,
				GetUserFn:      getUser,
			},
			WantOutput: "Profile 'bar' updated",
			ConfigPath: configPath,
			ConfigFile: &config.File{
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

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
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
