package profile_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	fsttime "github.com/fastly/cli/pkg/time"
)

func TestProfileCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate profile creation works",
			Args: "foo",
			API: mock.API{
				GetCurrentUserFn: getCurrentUser,
				GetTokenSelfFn:   getToken,
			},
			Stdin: []string{"some_token"},
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
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
			Args: "foo",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
					},
				},
			},
			WantError: "profile 'foo' already exists",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestProfileDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate profile deletion works",
			Args: "foo",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
					},
				},
			},
			WantOutput: "Profile 'foo' deleted",
		},
		{
			Name: "validate incorrect profile",
			Args: "unknown",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the specified profile does not exist",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestProfileList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate listing profiles works",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
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
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{},
			WantError:  "no profiles available",
		},
		{
			Name: "validate listing profiles with --verbose and --json causes an error",
			Args: "--verbose --json",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
					},
				},
			},
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate listing profiles with --json displays data correctly",
			Args: "--json",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			WantOutputs: []string{
				`"bar"`,
				`"token": "456"`,
				`"foo"`,
				`"token": "123"`,
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestProfileSwitch(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate switching to unknown profile returns an error",
			Args: "unknown",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the profile 'unknown' does not exist",
		},
		{
			Name: "validate switching profiles works",
			Args: "bar",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			WantOutput: "Profile switched to 'bar'",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "switch"}, scenarios)
}

func TestProfileToken(t *testing.T) {
	now := time.Now()
	expiredAt := now.Add(-600 * time.Second)
	soonExpireAt := now.Add(30 * time.Second)
	laterExpireAt := now.Add(1200 * time.Second)
	longTTLExpireAt := now.Add(60 * time.Second)

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate the active profile non-SSO token is displayed by default",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			WantOutput: "123",
		},
		{
			Name: "validate non-SSO token is displayed for the specified profile",
			Args: "bar",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			WantOutput: "456",
		},
		{
			Name: "validate non-SSO token is displayed for the specified profile using global --profile",
			Args: "--profile bar",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			WantOutput: "456",
		},
		{
			Name: "validate an unrecognised profile causes an error",
			Args: "unknown",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "profile 'unknown' does not exist",
		},
		{
			Name: "validate that an expired SSO token generates an error",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "123",
							Email:            "foo@example.com",
							RefreshExpiresAt: expiredAt.Format(time.RFC3339),
						},
					},
				},
			},
			WantError: fmt.Sprintf("the token in profile 'foo' expired at '%s'", expiredAt.UTC().Format(fsttime.Format)),
		},
		{
			Name: "validate that a soon-to-expire SSO token generates an error",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "123",
							Email:            "foo@example.com",
							RefreshExpiresAt: soonExpireAt.Format(time.RFC3339),
						},
					},
				},
			},
			WantError: fmt.Sprintf("the token in profile 'foo' will expire at '%s'", soonExpireAt.UTC().Format(fsttime.Format)),
		},
		{
			Name: "validate that a soon-to-expire SSO token with a non-default TTL does not generate an error",
			Args: "--ttl 30s",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "123",
							Email:            "foo@example.com",
							RefreshExpiresAt: longTTLExpireAt.Format(time.RFC3339),
						},
					},
				},
			},
			WantOutput: "123",
		},
		{
			Name: "validate that an SSO token with a long non-default TTL generates an error",
			Args: "--ttl 1800s",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "123",
							Email:            "foo@example.com",
							RefreshExpiresAt: laterExpireAt.Format(time.RFC3339),
						},
					},
				},
			},
			WantError: fmt.Sprintf("the token in profile 'foo' will expire at '%s'", laterExpireAt.UTC().Format(fsttime.Format)),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "token"}, scenarios)
}

func TestProfileUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate updating unknown profile returns an error",
			Args: "unknown",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantError: "the profile 'unknown' does not exist",
		},
		{
			Name: "validate updating profile works",
			Args: "bar",
			API: mock.API{
				GetCurrentUserFn: getCurrentUser,
				GetTokenSelfFn:   getToken,
			},
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "foo",
					Tokens: config.AuthTokens{
						"foo": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "123",
							Email: "foo@example.com",
						},
						"bar": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "456",
							Email: "bar@example.com",
						},
					},
				},
			},
			Stdin: []string{
				"",  // we skip updating the token
				"y", // we set the profile to be the default
			},
			WantOutput: "Profile 'bar' updated",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func getCurrentUser(_ context.Context) (*fastly.User, error) {
	return &fastly.User{
		Login:      fastly.ToPointer("foo@example.com"),
		CustomerID: fastly.ToPointer("abc"),
	}, nil
}

func getToken(_ context.Context) (*fastly.Token, error) {
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
