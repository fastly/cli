package testutil

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// MakeTempFile creates a tempfile with the given contents and returns its path.
func MakeTempFile(t *testing.T, contents string) string {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "fastly-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}

// CopyFile copies a referenced file to a new location.
func CopyFile(t *testing.T, fromFilename, toFilename string) {
	t.Helper()

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the variable.
	/* #nosec */
	src, err := os.Open(fromFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			t.Errorf("Failed to close fromFilename: %v", err)
		}
	}()

	toDir := filepath.Dir(toFilename)
	if err := os.MkdirAll(toDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the variable.
	/* #nosec */
	dst, err := os.Create(toFilename)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}

	if err := dst.Sync(); err != nil {
		t.Fatal(err)
	}

	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
}
