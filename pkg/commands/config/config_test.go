package config_test

import (
	"os"
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/testutil"
)

func TestConfig(t *testing.T) {
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

	scenarios := []testutil.TestScenario{
		{
			Name:       "validate config file content is displayed",
			ConfigPath: configPath,
			WantOutput: string(data),
		},
		{
			Name:       "validate config location is displayed",
			Arg:        "--location",
			ConfigPath: configPath,
			WantOutput: configPath,
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName}, scenarios)
}
