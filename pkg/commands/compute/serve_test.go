package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

// TestGetViceroy validates that Viceroy is installed to the appropriate
// directory.
//
// There isn't an executable binary that exists in the test environment, so we
// expect the spawning of a subprocess (to call `<binary> --version`) to error
// and subsequently the `installViceroy()` function to be called.
//
// The `installViceroy()` function will then think it has downloaded the latest
// release as we have instructed the mock to provide that behaviour.
//
// Subsequently the `os.Rename()` will move the downloaded Viceroy binary,
// which is just a dummy file created by `testutil.NewEnv`, into the intended
// destination directory.
func TestGetViceroy(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	viceroyBinName := "foo"
	installDirName := "install"

	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Dirs: []string{
			installDirName,
		},
		Write: []testutil.FileIO{
			{Src: "...", Dst: viceroyBinName},

			// NOTE: The reason for creating this file is that in serve.go when it tries
			// to write in-memory data back to disk, although we don't need to validate
			// the contents being written, we don't want the write to fail because no
			// such file existed.
			{Src: "", Dst: config.FileName},
		},
	})
	installDir := filepath.Join(rootdir, installDirName)
	binPath := filepath.Join(rootdir, viceroyBinName)
	configPath := filepath.Join(rootdir, config.FileName)
	defer os.RemoveAll(rootdir)

	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	compute.InstallDir = installDir

	var out bytes.Buffer

	av := mock.AssetVersioner{
		AssetVersion:   "1.2.3",
		BinaryFilename: viceroyBinName,
		DownloadOK:     true,
		DownloadedFile: binPath,
	}

	var file config.File

	// NOTE: We purposefully provide a nonsensical path, which we expect to fail,
	// but the function call should fallback to using the stubbed static config
	// defined above. We also don't pass stdin, stdout arguments as that
	// particular user flow isn't executed in this test case.
	err = file.Read("example", strings.NewReader("yes"), &out, fsterr.MockLog{}, false)
	if err != nil {
		t.Fatal(err)
	}

	g := global.Data{
		Config:     file,
		ConfigPath: configPath,
		ErrLog:     fsterr.MockLog{},
	}

	spinner, err := text.NewSpinner(&out)
	if err != nil {
		t.Fatal(err)
	}
	viceroyBinPath := "" // --viceroy-path flag for overriding CLI handling the Viceroy checks/downloads.
	viceroyCheck := false
	_, err = compute.GetViceroy(spinner, &out, av, &g, viceroyBinPath, viceroyCheck)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "Fetching latest Viceroy release") {
		t.Fatalf("expected file to be downloaded successfully")
	}

	movedPath := filepath.Join(installDir, viceroyBinName)

	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("binary was not moved to the install directory: %s", err)
	}
}
