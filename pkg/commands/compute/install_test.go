package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

// TestInstallTools validates that `compute install-tools` installs Viceroy
// to the appropriate directory using the same install path as `compute serve`.
//
// As with TestGetViceroy, there isn't an executable binary in the test
// environment, so the `<binary> --version` subprocess call errors and the
// installer downloads the (mocked) latest release, which `os.Rename()` then
// moves into the install directory.
func TestInstallTools(t *testing.T) {
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

			// NOTE: Created so the in-memory config can be written back to disk
			// without failing because no such file existed.
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
	defer func() {
		_ = os.Chdir(wd)
	}()

	github.InstallDir = installDir

	var out bytes.Buffer

	av := mock.AssetVersioner{
		AssetVersion:   "1.2.3",
		BinaryFilename: viceroyBinName,
		DownloadOK:     true,
		DownloadedFile: binPath,
	}

	var file config.File

	// NOTE: We purposefully provide a nonsensical path, which we expect to fail,
	// but the function call should fallback to using the stubbed static config.
	err = file.Read("example", strings.NewReader("yes"), &out, fsterr.MockLog{}, false)
	if err != nil {
		t.Fatal(err)
	}

	cmd := &compute.InstallCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				Config:     file,
				ConfigPath: configPath,
				ErrLog:     fsterr.MockLog{},
				Versioners: global.Versioners{
					Viceroy: av,
				},
			},
		},
	}
	if err := cmd.Exec(nil, &out); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "Fetching Viceroy release: ") {
		t.Fatalf("expected Viceroy to be downloaded successfully")
	}

	movedPath := filepath.Join(installDir, viceroyBinName)

	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("binary was not moved to the install directory: %s", err)
	}
}
