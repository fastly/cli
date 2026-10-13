package compute_test

import (
	"fmt"
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "package API error",
			Args: "-s 123 --version 1 --package pkg/package.tar.gz --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
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
			WantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
			WantOutputs: []string{
				"Uploading package",
			},
		},
		{
			Name: "success",
			Args: "-s 123 --version 2 --package pkg/package.tar.gz --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageOk,
			},
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
			WantOutputs: []string{
				"Uploading package",
				"Updated package (service 123, version 4)",
			},
		},
		{
			Name: "no fastly.toml manifest",
			Args: "-s 123 --version 2 --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			Env:             &testutil.EnvConfig{Opts: &testutil.EnvOpts{}},
			Setup:           readManifest,
			WantError:       "error reading fastly.toml: file not found",
			WantRemediation: "use the --package flag",
		},
		{
			Name: "fastly.toml missing name",
			Args: "-s 123 --version 2 --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
