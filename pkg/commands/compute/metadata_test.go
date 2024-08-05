package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml"

	root "github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/testutil"
)

func TestMetadata(t *testing.T) {
	// We read the static/embedded config so we can get the latest config
	// version and so we don't accidentally switch to the UseStatic() version.
	var staticConfig config.File
	err := toml.Unmarshal(config.Static, &staticConfig)
	if err != nil {
		t.Error(err)
	}

	scenarios := []testutil.TestScenario{
		{
			Arg: "--enable",
			ConfigFile: &config.File{
				ConfigVersion: staticConfig.ConfigVersion,
				CLI: config.CLI{
					Version: revision.SemVer(revision.AppVersion),
				},
			},
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "metadata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutput: "SUCCESS: configuration updated",
			Validator: func(t *testing.T, _ *testutil.TestScenario, opts *global.Data, _ error, _ bytes.Buffer) {
				data, err := os.ReadFile(opts.ConfigPath)
				if err != nil {
					t.Error(err)
				}

				var testFile config.File
				unmarshalErr := toml.Unmarshal(data, &testFile)
				if unmarshalErr != nil {
					t.Error(unmarshalErr)
				}

				testutil.AssertString(t, "enable", testFile.WasmMetadata.BuildInfo)
				testutil.AssertString(t, "enable", testFile.WasmMetadata.MachineInfo)
				testutil.AssertString(t, "enable", testFile.WasmMetadata.PackageInfo)
			},
		},
		{
			Arg: "--disable",
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "metadata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutput: "SUCCESS: configuration updated",
			Validator: func(t *testing.T, _ *testutil.TestScenario, opts *global.Data, _ error, _ bytes.Buffer) {
				data, err := os.ReadFile(opts.ConfigPath)
				if err != nil {
					t.Error(err)
				}

				var testFile config.File
				unmarshalErr := toml.Unmarshal(data, &testFile)
				if unmarshalErr != nil {
					t.Error(unmarshalErr)
				}

				testutil.AssertString(t, "disable", testFile.WasmMetadata.BuildInfo)
				testutil.AssertString(t, "disable", testFile.WasmMetadata.MachineInfo)
				testutil.AssertString(t, "disable", testFile.WasmMetadata.PackageInfo)
			},
		},
		{
			Arg: "--enable --disable-build",
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "metadata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutputs: []string{
				"INFO: We will enable all metadata except for the specified `--disable-*` flags",
				"SUCCESS: configuration updated",
			},
			Validator: func(t *testing.T, _ *testutil.TestScenario, opts *global.Data, _ error, _ bytes.Buffer) {
				data, err := os.ReadFile(opts.ConfigPath)
				if err != nil {
					t.Error(err)
				}

				var testFile config.File
				unmarshalErr := toml.Unmarshal(data, &testFile)
				if unmarshalErr != nil {
					t.Error(unmarshalErr)
				}

				testutil.AssertString(t, "disable", testFile.WasmMetadata.BuildInfo)
				testutil.AssertString(t, "enable", testFile.WasmMetadata.MachineInfo)
				testutil.AssertString(t, "enable", testFile.WasmMetadata.PackageInfo)
			},
		},
		{
			Arg: "--disable --enable-machine",
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "metadata", "config.toml"),
							Dst: "config.toml",
						},
					},
				},
				EditScenario: func(scenario *testutil.TestScenario, rootdir string) {
					scenario.ConfigPath = filepath.Join(rootdir, "config.toml")
				},
			},
			WantOutputs: []string{
				"INFO: We will disable all metadata except for the specified `--enable-*` flags",
				"SUCCESS: configuration updated",
			},
			Validator: func(t *testing.T, _ *testutil.TestScenario, opts *global.Data, _ error, _ bytes.Buffer) {
				data, err := os.ReadFile(opts.ConfigPath)
				if err != nil {
					t.Error(err)
				}

				var testFile config.File
				unmarshalErr := toml.Unmarshal(data, &testFile)
				if unmarshalErr != nil {
					t.Error(unmarshalErr)
				}

				testutil.AssertString(t, "disable", testFile.WasmMetadata.BuildInfo)
				testutil.AssertString(t, "enable", testFile.WasmMetadata.MachineInfo)
				testutil.AssertString(t, "disable", testFile.WasmMetadata.PackageInfo)
			},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "metadata"}, scenarios)
}
