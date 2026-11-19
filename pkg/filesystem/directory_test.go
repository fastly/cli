package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFileNonRegularDestinationErrorNamesDestination(t *testing.T) {
	tmp := t.TempDir()

	src := filepath.Join(tmp, "source.txt")
	if err := os.WriteFile(src, []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// A directory is not a regular file, so CopyFile must reject it as a
	// destination. The resulting error must name the destination path, not the
	// source path, so users can identify which path caused the failure.
	dst := filepath.Join(tmp, "destination-dir")
	if err := os.Mkdir(dst, 0o700); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	err := CopyFile(src, dst)
	if err == nil {
		t.Fatal("expected an error when copying to a non-regular destination")
	}

	if !strings.Contains(err.Error(), "non-regular destination file") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
	if !strings.Contains(err.Error(), dst) {
		t.Errorf("error %q does not name the destination path %q", err.Error(), dst)
	}
	if strings.Contains(err.Error(), src) {
		t.Errorf("error %q incorrectly names the source path %q", err.Error(), src)
	}
}
