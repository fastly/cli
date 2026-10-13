package compute_test

import (
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

// readManifest replaces the manifest read by testutil.MockGlobalData, which
// happens before the scenario runner changes into the test environment, with
// one read from the test environment (as app.Init would in production).
func readManifest(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
	var md manifest.Data
	md.File.Args = opts.Args
	md.File.SetErrLog(opts.ErrLog)
	md.File.SetOutput(opts.Output)
	_ = md.File.Read(manifest.Filename)
	opts.Manifest = &md
}

func TestValidate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "success",
			Args: "--package pkg/package.tar.gz",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Copy: []testutil.FileIO{
						{
							Src: filepath.Join("testdata", "deploy", "pkg", "package.tar.gz"),
							Dst: filepath.Join("pkg", "package.tar.gz"),
						},
					},
				},
			},
			WantError:  "",
			WantOutput: "Validated package",
		},
		{
			Name:            "no fastly.toml manifest",
			Env:             &testutil.EnvConfig{Opts: &testutil.EnvOpts{}},
			Setup:           readManifest,
			WantError:       "error reading fastly.toml: file not found",
			WantRemediation: "use the --package flag",
		},
		{
			Name: "fastly.toml missing name",
			Env: &testutil.EnvConfig{
				Opts: &testutil.EnvOpts{
					Write: []testutil.FileIO{
						{Src: "manifest_version = 3\n", Dst: manifest.Filename},
					},
				},
			},
			Setup:     readManifest,
			WantError: "no name found in the fastly.toml",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "validate"}, scenarios)
}
