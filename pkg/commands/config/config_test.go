package config_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
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

	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:       "validate config file content is displayed",
			Args:       args("config"),
			WantOutput: string(data),
		},
		{
			Name:       "validate config location is displayed",
			Args:       args("config --location"),
			WantOutput: configPath,
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				opts.ConfigPath = configPath
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
