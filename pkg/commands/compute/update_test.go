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
	scenarios := []testutil.TestScenario{
		{
			Name: "package API error",
			Arg:  "-s 123 --version 1 --package pkg/package.tar.gz --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
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
			Arg:  "-s 123 --version 2 --package pkg/package.tar.gz --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageOk,
			},
			NewEnv: &testutil.NewEnvConfig{
				EnvOpts: &testutil.EnvOpts{
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

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
