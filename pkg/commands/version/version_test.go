package version_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersion(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("skipping test due to unix specific mock shell script")
	}

	// We're going to chdir to an temp environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Write: []testutil.FileIO{
			{Src: `#!/bin/bash
			echo viceroy 0.0.0`, Dst: "viceroy"},
		},
	})
	defer os.RemoveAll(rootdir)

	// Ensure the viceroy file we created can be executed.
	//
	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as this is for test suite purposes only.
	// #nosec
	err = os.Chmod(filepath.Join(rootdir, "viceroy"), 0o777)
	if err != nil {
		t.Fatal(err)
	}

	// Override the InstallDir where the viceroy binary is looked up.
	orgInstallDir := github.InstallDir
	github.InstallDir = rootdir
	defer func() {
		github.InstallDir = orgInstallDir
	}()

	// Before running the test, chdir into the temp environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably assert file structure.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(pwd)
	}()

	var stdout bytes.Buffer
	args := testutil.Args("version")
	opts := testutil.MockGlobalData(args, &stdout)
	opts.Versioners = global.Versioners{
		Viceroy: github.New(github.Opts{
			Org:    "fastly",
			Repo:   "viceroy",
			Binary: "viceroy",
		}),
	}
	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		return opts, nil
	}
	err = app.Run(args, nil)

	t.Log(stdout.String())

	testutil.AssertNoError(t, err)
	testutil.AssertString(t, strings.Join([]string{
		"Fastly CLI version v0.0.0-unknown (unknown)",
		fmt.Sprintf("Built with go version %s unknown/unknown", runtime.Version()),
		"Viceroy version: viceroy 0.0.0",
		"",
	}, "\n"), stdout.String())
}
