package profile_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestProfileCreate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate profile creation works",
			Arg:  "foo",
			API: mock.API{
				GetTokenSelfFn: getToken,
				GetUserFn:      getUser,
			},
			Stdin: []string{"some_token"},
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutputs: []string{
				"Fastly API token:",
				"Validating token",
				"Persisting configuration",
				"Profile 'foo' created",
			},
		},
		{
			Name: "validate profile duplication",
			Arg:  "foo",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
				},
			},
			WantError: "profile 'foo' already exists",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestProfileDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate profile deletion works",
			Arg:  "foo",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"foo": &config.Profile{
						Default: true,
						Email:   "foo@example.com",
						Token:   "123",
					},
				},
			},
			WantOutput: "Profile 'foo' deleted",
		},
		{
			Name: "validate incorrect profile",
			Arg:  "unknown",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the specified profile does not exist",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestProfileList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate listing profiles works",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutputs: []string{
				"Default profile highlighted in red.",
				"foo\n\nDefault: true\nEmail: foo@example.com\nToken: 123",
				"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
			},
		},
		{
			Name: "validate no profiles defined",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{},
			WantError:  "no profiles available",
		},
		// NOTE: The following test is subtly different to the previous one in that
		// our logic checks whether the config.Profiles map type is nil. If it is
		// then we error (see above test), otherwise if the map is set but there
		// are no profiles, then we notify the user no profiles exist.
		{
			Name: "validate no profiles available",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{},
			},
			WantOutputs: []string{
				"No profiles defined. To create a profile, run",
				"fastly profile create <name>",
			},
		},
		{
			Name: "validate listing profiles displays warning if no default set",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutputs: []string{
				"At least one account profile should be set as the 'default'.",
				"foo\n\nDefault: false\nEmail: foo@example.com\nToken: 123",
				"bar\n\nDefault: false\nEmail: bar@example.com\nToken: 456",
			},
		},
		{
			Name: "validate listing profiles with --verbose and --json causes an error",
			Arg:  "--verbose --json",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate listing profiles with --json displays data correctly",
			Arg:  "--json",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestProfileSwitch(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate switching to unknown profile returns an error",
			Arg:  "unknown",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the profile 'unknown' does not exist",
		},
		{
			Name: "validate switching profiles works",
			Arg:  "bar",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutput: "Profile switched to 'bar'",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "switch"}, scenarios)
}

func TestProfileToken(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate the active profile token is displayed by default",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutput: "123",
		},
		{
			Name: "validate token is displayed for the specified profile",
			Arg:  "bar", // we choose a non-default profile
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutput: "456",
		},
		{
			Name: "validate token is displayed for the specified profile using global --profile",
			Arg:  "--profile bar", // we choose a non-default profile
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutput: "456",
		},
		{
			Name: "validate an unrecognised profile causes an error",
			Arg:  "unknown",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "profile 'unknown' does not exist",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "token"}, scenarios)
}

func TestProfileUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate updating unknown profile returns an error",
			Arg:  "unknown",
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the profile 'unknown' does not exist",
		},
		{
			Name: "validate updating profile works",
			Arg:  "bar", // we choose a non-default profile
			API: mock.API{
				GetTokenSelfFn: getToken,
				GetUserFn:      getUser,
			},
			Env: &testutil.EnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
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
			WantOutput: "Profile 'bar' updated",
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
