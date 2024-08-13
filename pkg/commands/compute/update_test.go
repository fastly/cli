package compute_test

import (
	"fmt"
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "package API error",
			Args: "-s 123 --version 1 --package pkg/package.tar.gz --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
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
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
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
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
