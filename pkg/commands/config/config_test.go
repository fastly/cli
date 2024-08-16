package config_test

import (
	"os"
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/testutil"
)

func TestConfig(t *testing.T) {
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

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate config file content is displayed",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Write: []testutil.FileIO{
						{Src: string(data), Dst: "config.toml"},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutput: string(data),
		},
		{
			Name: "validate config location is displayed",
			Args: "--location",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Write: []testutil.FileIO{
						{Src: string(data), Dst: "config.toml"},
					},
				},
				EditScenario: func(scenario *testutil.CLIScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
					scenario.WantOutput = scenario.ConfigPath
				},
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName}, scenarios)
}
