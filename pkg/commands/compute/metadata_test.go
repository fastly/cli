package compute_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/testutil"
)

// Scenario is an extension of the base TestScenario.
// It includes manipulating stdin.
type Scenario struct {
	testutil.TestScenario

	ExpectedConfig config.WasmMetadata
}

func TestMetadata(t *testing.T) {
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
		path, err := filepath.Abs(filepath.Join("./", "testdata", "metadata", "config.toml"))
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
				Args:       args("compute metadata --enable"),
				WantOutput: "SUCCESS: configuration updated (see: `fastly config`)",
			},
			ExpectedConfig: config.WasmMetadata{
				BuildInfo:   "enable",
				MachineInfo: "enable",
				PackageInfo: "enable",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args:       args("compute metadata --disable"),
				WantOutput: "SUCCESS: configuration updated (see: `fastly config`)",
			},
			ExpectedConfig: config.WasmMetadata{
				BuildInfo:   "disable",
				MachineInfo: "disable",
				PackageInfo: "disable",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: args("compute metadata --enable --disable-build"),
				WantOutputs: []string{
					"INFO: We will enable all metadata except for the specified `--disable-*` flags",
					"SUCCESS: configuration updated (see: `fastly config`)",
				},
			},
			ExpectedConfig: config.WasmMetadata{
				BuildInfo:   "disable",
				MachineInfo: "enable",
				PackageInfo: "enable",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: args("compute metadata --disable --enable-machine"),
				WantOutputs: []string{
					"INFO: We will disable all metadata except for the specified `--enable-*` flags",
					"SUCCESS: configuration updated (see: `fastly config`)",
				},
			},
			ExpectedConfig: config.WasmMetadata{
				BuildInfo:   "disable",
				MachineInfo: "enable",
				PackageInfo: "disable",
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer

			opts := testutil.MockGlobalData(testcase.Args, &stdout)

			// We override the config path so that we don't accidentally write over
			// our own configuration file.
			opts.ConfigPath = configPath

			// We read the static/embedded config so we can get the latest config
			// version and so we don't accidentally switch to the UseStatic() version.
			var staticConfig config.File
			err := toml.Unmarshal(config.Static, &staticConfig)
			if err != nil {
				t.Error(err)
			}

			// The read of the config file only happens in the main() function, so for
			// the sake of the test environment we need to construct an in-memory
			// representation of the config file we want to be using.
			opts.Config = config.File{
				ConfigVersion: staticConfig.ConfigVersion,
				CLI: config.CLI{
					Version: revision.SemVer(revision.AppVersion),
				},
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
			for _, s := range testcase.WantOutputs {
				testutil.AssertStringContains(t, stdout.String(), s)
			}

			in := strings.NewReader("")
			verboseMode := false
			err = opts.Config.Read(configPath, in, opts.Output, opts.ErrLog, verboseMode)
			if err != nil {
				t.Error(err)
			}

			if opts.Config.WasmMetadata.BuildInfo != testcase.ExpectedConfig.BuildInfo {
				t.Errorf("want: %s, got: %s", testcase.ExpectedConfig.BuildInfo, opts.Config.WasmMetadata.BuildInfo)
			}
			if opts.Config.WasmMetadata.MachineInfo != testcase.ExpectedConfig.MachineInfo {
				t.Errorf("want: %s, got: %s", testcase.ExpectedConfig.MachineInfo, opts.Config.WasmMetadata.MachineInfo)
			}
			if opts.Config.WasmMetadata.PackageInfo != testcase.ExpectedConfig.PackageInfo {
				t.Errorf("want: %s, got: %s", testcase.ExpectedConfig.PackageInfo, opts.Config.WasmMetadata.PackageInfo)
			}
		})
	}
}
