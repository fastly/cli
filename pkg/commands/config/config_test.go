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

	scenarios := []testutil.TestScenario{
		{
			Name: "validate config file content is displayed",
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Write: []testutil.FileIO{
						{Src: string(data), Dst: "config.toml"},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutput: string(data),
		},
		{
			Name: "validate config location is displayed",
			Arg:  "--location",
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Write: []testutil.FileIO{
						{Src: string(data), Dst: "config.toml"},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
					scenario.WantOutput = scenario.ConfigPath
				},
			},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName}, scenarios)
}
