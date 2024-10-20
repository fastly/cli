package compute_test

import (
	"path/filepath"
	"testing"

	root "github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/testutil"
)

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
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "validate"}, scenarios)
}
