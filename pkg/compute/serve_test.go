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

func TestGetViceroy(t *testing.T) {
	binary := "foo"
	dir, downloadedFile := makeEnvironment(binary, t)

	defer os.RemoveAll(dir) // clean up

	InstallDir = dir

	var out bytes.Buffer

	progress := text.NewQuietProgress(&out)
	versioner := mock.Versioner{
		Version:        "v1.2.3",
		Binary:         binary,
		DownloadOK:     true,
		DownloadedFile: downloadedFile,
	}

	// There isn't an executable binary that exists in the test environment, so
	// we expect the spawning of a subprocess to call `<binary> --version` to
	// fail and subsequently the `installViceroy()` function to be called. This
	// function will then think it has downloaded the latest release (as we mock
	// that behaviour) and so we should see the correct output.
	bin, err := getViceroy(progress, &out, versioner)
	if err != nil {
		t.Fatal(err)
	}

	if bin != downloadedFile {
		t.Fatalf("want: %s, have: %s", downloadedFile, bin)
	}

	if !strings.Contains(out.String(), "✓ Fetching latest viceroy release") {
		t.Fatalf("expected file to be downloaded successfully")
	}
}

func makeEnvironment(downloadedFilename string, t *testing.T) (string, string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-serve-*")
	if err != nil {
		t.Fatal(err)
	}

	fpath := filepath.Join(rootdir, downloadedFilename)
	if err := os.WriteFile(fpath, []byte("..."), 0777); err != nil {
		t.Fatal(err)
	}

	return rootdir, fpath
}
