package compute

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/text"
)

// TestGetViceroy validates that viceroy is installed to the appropriate
// directory.
//
// There isn't an executable binary that exists in the test environment, so we
// expect the spawning of a subprocess to call `<binary> --version` to fail and
// subsequently the `installViceroy()` function to be called.
//
// The `installViceroy()` function will then think it has downloaded the latest
// release as we have instructed the mock to provide that behaviour.
//
// Subsequently the `os.Rename()` will move the downloaded viceroy binary,
// which is just a dummy file created by `makeEnvironment()`, into the intended
// destination directory. The destination directory in the case of the test
// environment is ...
func TestGetViceroy(t *testing.T) {
	binary := "foo"
	downloadDir, installDir, downloadedFile := makeEnvironment(binary, t)

	defer os.RemoveAll(downloadDir) // clean up

	InstallDir = installDir

	var out bytes.Buffer

	progress := text.NewQuietProgress(&out)
	versioner := mock.Versioner{
		Version:        "v1.2.3",
		Binary:         binary,
		DownloadOK:     true,
		DownloadedFile: downloadedFile,
	}

	_, err := getViceroy(progress, &out, versioner)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "✓ Fetching latest viceroy release") {
		t.Fatalf("expected file to be downloaded successfully")
	}

	movedPath := filepath.Join(installDir, binary)

	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("binary was not moved to the install directory: %s", err)
	}
}

// makeEnvironment creates a temporary directory for the test suite to utilise
// when validating viceroy installation behaviours.
//
// It will create a file within the temporary directory to represent the call
// to `versioner.Download()` being successful.
//
// It also creates a nested directory within the temp directory to represent
// where the downloaded binary should be moved into.
func makeEnvironment(downloadedFilename string, t *testing.T) (string, string, string) {
	t.Helper()

	downloadDir, err := os.MkdirTemp("", "fastly-serve-*")
	if err != nil {
		t.Fatal(err)
	}

	fpath := filepath.Join(downloadDir, downloadedFilename)
	if err := os.WriteFile(fpath, []byte("..."), 0777); err != nil {
		t.Fatal(err)
	}

	installDir, err := os.MkdirTemp(downloadDir, "install")
	if err != nil {
		t.Fatal(err)
	}

	return downloadDir, installDir, fpath
}
