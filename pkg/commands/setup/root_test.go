package setup_test

import (
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/testutil"
)

func TestSetupNonInteractive(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "non-interactive requires token",
			Args: "--non-interactive",
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
			WantError: "--token is required when using --non-interactive",
		},
		{
			Name: "non-interactive profile already exists",
			Args: "--non-interactive --token test_token_123 --name user",
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
				Profiles: config.Profiles{
					"user": &config.Profile{
						Default: true,
						Email:   "existing@example.com",
						Token:   "existing_token",
					},
				},
			},
			WantError: "profile 'user' already exists",
		},
		{
			Name: "auto-yes profile already exists",
			Args: "--auto-yes --name user",
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
				Profiles: config.Profiles{
					"user": &config.Profile{
						Default: true,
						Email:   "existing@example.com",
						Token:   "existing_token",
					},
				},
			},
			WantError: "profile 'user' already exists",
		},
	}

	testutil.RunCLIScenarios(t, []string{"setup"}, scenarios)
}
